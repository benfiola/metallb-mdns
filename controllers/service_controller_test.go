package controllers

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func createService() *corev1.Service {
	ctx := context.Background()

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-service",
			Namespace: "default",
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{},
			Ports: []corev1.ServicePort{{
				Port: 80,
				TargetPort: intstr.IntOrString{IntVal: 80},
			}},
		},
	}
	Expect(k8sClient.Create(ctx, service)).Should(Succeed())
	return service
}

var _ = Describe("Ingress Controller", func() {
	BeforeEach(func() {

	})

	AfterEach(func() {

	})

	Context("Functionality", func() {
		It("Should ignore new services with an empty/missing hostname annotation", func() {

		})

		It("Should ignore new services without an external IP", func() {

		})

		It("Should clean-up services that change to an empty/missing hostname annotation", func() {

		})

		It("Should clean-up services that change to no longer have an external IP", func() {

		})

		It("Should create host entries for new services", func() {

		})

		It("Should update host entries for existing services that change", func() {

		})
	})
})
