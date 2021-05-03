package reconciler

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"

	v1alpha1 "github.com/cirocosta/monero-operator/pkg/apis/utxo.com.br/v1alpha1"
)

func AppLabel(name string) map[string]string {
	return map[string]string{
		"app": name,
	}
}

func NewMonerodContainer(nodeSet *v1alpha1.MoneroNodeSet) corev1.Container {
	defaultArgs := []string{
		"--data-dir=" + MonerodDataVolumeMountPath,

		"--non-interactive",
		"--no-zmq",
		"--no-igd",

		"--p2p-bind-ip=0.0.0.0",
		"--p2p-bind-port=18080",

		"--rpc-restricted-bind-ip=0.0.0.0",
		"--rpc-restricted-bind-port=18089",
	}

	command := append([]string{
		"monerod",
	}, MergedSlice(defaultArgs, nodeSet.Spec.Monerod.Args)...)

	obj := corev1.Container{
		Name:    MonerodContainerName,
		Image:   MonerodContainerImage,
		Command: command,
		ReadinessProbe: &corev1.Probe{
			PeriodSeconds:       15,
			InitialDelaySeconds: 15,
			FailureThreshold:    5,
			Handler: corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: MonerodContainerProbePath,
					Port: intstr.FromString(MonerodContainerProbePort),
				},
			},
		},
		Ports: []corev1.ContainerPort{
			{
				Name:          P2PPortName,
				ContainerPort: int32(P2PPortNumber),
				Protocol:      corev1.ProtocolTCP,
			},

			{
				Name:          RestrictedPortName,
				ContainerPort: int32(RestrictedPortNumber),
				Protocol:      corev1.ProtocolTCP,
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      MonerodDataVolumeName,
				MountPath: MonerodDataVolumeMountPath,
			},
		},
	}

	return obj
}

func NewPodTemplateSpec(nodeSet *v1alpha1.MoneroNodeSet) corev1.PodTemplateSpec {
	o := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: AppLabel(nodeSet.Name),
		},
		Spec: corev1.PodSpec{
			TerminationGracePeriodSeconds: pointer.Int64Ptr(60),
			Containers: []corev1.Container{
				NewMonerodContainer(nodeSet),
			},
		},
	}

	if nodeSet.Spec.HardAntiAffinity == true {
		o.Spec.Affinity = &corev1.Affinity{
			PodAntiAffinity: &corev1.PodAntiAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
					{
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: AppLabel(nodeSet.Name),
						},
					},
				},
			},
		}
	}

	return o
}

func NewVolumeClaimTemplate(nodeSet *v1alpha1.MoneroNodeSet) corev1.PersistentVolumeClaim {
	o := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name: MonerodDataVolumeName,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(nodeSet.Spec.DiskSize),
				},
			},
		},
	}

	if nodeSet.Spec.StorageClass != "" {
		o.Spec.StorageClassName = pointer.StringPtr(nodeSet.Spec.StorageClass)
	}

	return o
}

func NewStatefulSet(nodeSet *v1alpha1.MoneroNodeSet) *appsv1.StatefulSet {
	obj := &appsv1.StatefulSet{}

	obj.TypeMeta = metav1.TypeMeta{
		Kind:       "StatefulSet",
		APIVersion: appsv1.SchemeGroupVersion.Identifier(),
	}

	obj.ObjectMeta = metav1.ObjectMeta{
		Name:      nodeSet.Name,
		Namespace: nodeSet.Namespace,
	}

	obj.Spec = appsv1.StatefulSetSpec{
		ServiceName: nodeSet.Name,
		Replicas:    pointer.Int32Ptr(int32(nodeSet.Spec.Replicas)),

		RevisionHistoryLimit: pointer.Int32Ptr(0),
		Selector: &metav1.LabelSelector{
			MatchLabels: AppLabel(nodeSet.Name),
		},

		Template: NewPodTemplateSpec(nodeSet),
		VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
			NewVolumeClaimTemplate(nodeSet),
		},
	}

	return obj
}
