# metallb-mdns

Implementation of a kubernetes `Service` reflector that synchronizes `Service` resources annotated with
 `metallb-mdns/hostname` that have external IPs allocated with the control plane's `/etc/avahi/hosts` file.

By doing so, services exposed via `LoadBalancer` on cluster can be reached over mDNS on the host local network.

The scaffolding for the project was initially generated using [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder).

## Requirements

* Working kubernetes cluster
* `avahi-daemon` running on the Kubernetes control plane
* The ability to deploy workloads to the control plane
* The ability to mount `/etc/avahi/hosts` with read/write access to a pod running on the control plane

## Deployment

Here's an [example](https://raw.githubusercontent.com/benfiola/metallb-mdns/main/config/example.yaml) manifest that can be deployed with:

```shell
kubectl apply -f https://raw.githubusercontent.com/benfiola/metallb-mdns/main/config/example.yaml
```

This manifest contains:

* Namespace
* Service Account
* Cluster Role (granting the ability to monitor + edit `Service` resources)
* Cluster Role Binding (connecting the `Service Account` to the `Cluster Role`)
* Deployment (a controller monitoring `Service` resources) - this has to be run on the control plane.  This deployment has a `hostsPath` volume mount providing the control plane's Avahi `/etc/avahi/hosts` path to the running pod.

## Example

Here's an example of a `Service` annotated to define an mDNS hostname:

```yaml
apiVersion: v1
kind: Service
metadata:
    name: example-service
    annotations:
      "metallb-mdns/hostname": service.local
spec:
  ...
  type: LoadBalancer
```

## TODO

1. Gracefully handle pod termination (cleanup would allow the controller to run on any node)
2. Continued removal of unnecessary scaffolding generated from kubebuilder
3. Reimplementation without kubebuilder (?) (without any CRDs, a lot of this scaffolding is unnecessary)
4. Implementation of basic unit tests