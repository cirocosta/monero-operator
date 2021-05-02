package reconciler

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func NewService(name, namespace string) *corev1.Service {
	obj := &corev1.Service{}

	obj.TypeMeta = metav1.TypeMeta{
		Kind:       "Service",
		APIVersion: corev1.SchemeGroupVersion.Identifier(),
	}

	obj.ObjectMeta = metav1.ObjectMeta{
		Name:      name,
		Namespace: namespace,
		Labels:    AppLabel(name),
	}

	obj.Spec = corev1.ServiceSpec{
		Selector: AppLabel(name),
		Ports: []corev1.ServicePort{
			{
				Name:       P2PPortName,
				Port:       int32(P2PPortNumber),
				TargetPort: intstr.FromInt(int(P2PPortNumber)),
				Protocol:   corev1.ProtocolTCP,
			},

			{
				Name:       RestrictedPortName,
				Port:       int32(RestrictedPortNumber),
				TargetPort: intstr.FromInt(int(RestrictedPortNumber)),
				Protocol:   corev1.ProtocolTCP,
			},
		},
	}

	return obj
}
