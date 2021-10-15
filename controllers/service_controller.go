/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"bufio"
	"context"
	"fmt"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"os"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"strings"
)

// ServiceReconciler reconciles a Service object
type ServiceReconciler struct {
	client.Client
	Logger logr.Logger
	Scheme *runtime.Scheme
	HostsFile string
}

const (
	baseKey = "metallb-mdns"
	hostnameKey = baseKey + "/hostname"
	processedHostnameKey = baseKey + "/processedHostname"
	finalizerKey = baseKey + "/finalizer"
)

func (r *ServiceReconciler) GetHosts() map[string]string {
	// open file for writing
	file, err := os.Open(r.HostsFile)
	if err != nil {
		panic(err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	// read hosts file line-by-line
	hostsMap := map[string]string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		// ignore empty lines
		if len(line) == 0 {
			continue
		}
		// ignore comments
		if line[0] == '#' {
			continue
		}

		// ignore lines that do not have <ip> <host> segments
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		// create hostname -> ipaddress mapping
		hostsMap[parts[1]] = parts[0]
	}

	// propagate read failures
	if err := scanner.Err(); err != nil {
		panic(err)
	}

	return hostsMap
}

func (r *ServiceReconciler) SetHosts(hosts map[string]string) {
	// only write to disk if hosts have changed
	oldHosts := r.GetHosts()
	if reflect.DeepEqual(oldHosts, hosts) {
		return
	}

	r.Logger.Info(fmt.Sprintf("saving: %s", r.HostsFile))

	// open file for writing
	file, err := os.Create(r.HostsFile)
	if err != nil {
		panic(err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	// write to file with format:
	//   <ipaddress> <host>
	for hostname, ipAddress := range hosts {
		_, err := file.WriteString(fmt.Sprintf("%s %s\n", ipAddress, hostname))
		if err != nil {
			panic(err)
		}
	}
}

func (r *ServiceReconciler) AddFinalizer(service *corev1.Service) {
	if ! controllerutil.ContainsFinalizer(service, finalizerKey) {
		r.Logger.Info("adding finalizer")
		controllerutil.AddFinalizer(service, finalizerKey)
	}
}

func (r *ServiceReconciler) RemoveFinalizer(service *corev1.Service) {
	if controllerutil.ContainsFinalizer(service, finalizerKey) {
		r.Logger.Info("removing finalizer")
		controllerutil.RemoveFinalizer(service, finalizerKey)
	}
}

func (r *ServiceReconciler) Save(ctx context.Context, service *corev1.Service) (ctrl.Result, error) {
	/* syncs a local kubernetes resource with the kubernetes api */
	annotations := service.GetAnnotations()
	_, hasAnnotation := annotations[processedHostnameKey]

	// determine whether to add finalizers based upon presence
	// of controller annotation
	if hasAnnotation {
		r.AddFinalizer(service)
	} else {
		r.RemoveFinalizer(service)
	}

	err := r.Update(ctx, service)
	return ctrl.Result{}, err
}

func (r *ServiceReconciler) Cleanup(ctx context.Context, service *corev1.Service) (ctrl.Result, error) {
	/* called when a service is no longer managed by the controller */
	r.Logger.Info("cleanup")

	annotations := service.GetAnnotations()
	hostname := annotations[processedHostnameKey]
	hosts := r.GetHosts()

	// clear hostnames
	delete(hosts, hostname)
	r.SetHosts(hosts)

	// clear controller-set annotations
	delete(annotations, processedHostnameKey)

	// save
	return r.Save(ctx, service)
}

func (r *ServiceReconciler) OnChange(ctx context.Context, service *corev1.Service) (ctrl.Result, error) {
	/* called on every service change */
	r.Logger.Info("change")

	// ensure annotations is set to a map
	// (this is only saved if a relevant change occurs)
	annotations := service.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
		service.SetAnnotations(annotations)
	}

	hosts := r.GetHosts()
	hostname := annotations[processedHostnameKey]
	externalIP := hosts[hostname]

	// handle case when hostname changes
	if hostname != annotations[hostnameKey] {
		newHostname := annotations[hostnameKey]

		// handle case if hostname is unset
		if newHostname == "" {
			r.Logger.Info("hostname unset")
			return r.Cleanup(ctx, service)
		}

		// delete old hosts entry + save
		// update annotations
		// NOTE: upon save, onChange is triggered (again) - this time, syncing the ip address with the new hostname.
		r.Logger.Info(fmt.Sprintf("hostname change: %s", newHostname))
		delete(hosts, hostname)
		r.SetHosts(hosts)
		annotations[processedHostnameKey] = newHostname
		return r.Save(ctx, service)
	}

	newExternalIP := ""
	if len(service.Status.LoadBalancer.Ingress) > 0 {
		newExternalIP = service.Status.LoadBalancer.Ingress[0].IP
	}

	// handle case when ip address changes
	if hostname != "" && externalIP != newExternalIP {
		// handle case when ip address is unset
		if newExternalIP == "" {
			r.Logger.Info("ip address unset")
			return r.Cleanup(ctx, service)
		}

		// update hosts entry to use new ip address
		r.Logger.Info(fmt.Sprintf("ip address change: %s", newExternalIP))
		hosts[hostname] = newExternalIP
		r.SetHosts(hosts)
		return r.Save(ctx, service)
	}

	// handle case
	// no meaningful changes - no-op
	return ctrl.Result{}, nil
}


//+kubebuilder:rbac:groups=,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=,resources=services/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=,resources=services/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *ServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Logger = log.FromContext(ctx)

	// get service
	var service corev1.Service
	if err := r.Get(ctx, req.NamespacedName, &service); err != nil {
		r.Logger.Info("service not found")
		return ctrl.Result{}, nil
	}

	if ! service.ObjectMeta.GetDeletionTimestamp().IsZero() {
		return r.Cleanup(ctx, &service)
	}
	return r.OnChange(ctx, &service)
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Logger = log.FromContext(nil)
	r.Logger.Info("controller setup")

	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Service{}).
		Complete(r)
}
