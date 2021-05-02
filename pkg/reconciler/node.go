package reconciler

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	P2PPortName          = "p2p"
	P2PPortNumber uint16 = 18080

	RestrictedPortName          = "restricted"
	RestrictedPortNumber uint16 = 18089

	MonerodContainerName      = "monerod"
	MonerodContainerImage     = "index.docker.io/utxobr/monerod@sha256:19ba5793c00375e7115469de9c14fcad928df5867c76ab5de099e83f646e175d"
	MonerodContainerProbePath = "/get_info"
	MonerodContainerProbePort = RestrictedPortName

	MonerodDataVolumeName      = "data"
	MonerodDataVolumeMountPath = "/data"
)

func NewService(name, namespace string) *corev1.Service {
	obj := &corev1.Service{}

	obj.ObjectMeta = metav1.ObjectMeta{
		Name:      name,
		Namespace: namespace,
		Labels:    AppLabel(name),
	}

	obj.Spec = corev1.ServiceSpec{
		Selector: map[string]string{},

		Ports: []corev1.ServicePort{
			{
				Name:     P2PPortName,
				Port:     int32(P2PPortNumber),
				Protocol: corev1.ProtocolTCP,
			},

			{
				Name:     RestrictedPortName,
				Port:     int32(RestrictedPortNumber),
				Protocol: corev1.ProtocolTCP,
			},
		},
	}

	return obj
}

func AppLabel(name string) map[string]string {
	return map[string]string{
		"app": name,
	}
}

func NewMonerodContainer() corev1.Container {
	obj := corev1.Container{
		Name:  MonerodContainerName,
		Image: MonerodContainerImage,
		Command: []string{
			"monerod",
			"--data-dir=" + MonerodDataVolumeMountPath,
			"--enable-dns-blocklist",
			"--no-igd",
			"--non-interactive",
			"--p2p-bind-ip=0.0.0.0",
			"--p2p-bind-port=18080",
			"--rpc-restricted-bind-ip=0.0.0.0",
			"--rpc-restricted-bind-port=18089",
			"--testnet",
		},
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

func NewStatefulSet(name, namespace string) *appsv1.StatefulSet {
	obj := &appsv1.StatefulSet{}

	obj.ObjectMeta = metav1.ObjectMeta{
		Name:      name,
		Namespace: namespace,
	}

	obj.Spec = appsv1.StatefulSetSpec{
		ServiceName: name,

		Replicas: int32p(1),

		RevisionHistoryLimit: int32p(0),

		Selector: &metav1.LabelSelector{
			MatchLabels: AppLabel(name),
		},

		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: AppLabel(name),
			},
			Spec: corev1.PodSpec{
				TerminationGracePeriodSeconds: int64p(60),
				Containers: []corev1.Container{
					NewMonerodContainer(),
				},
			},
		},
		VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: MonerodDataVolumeName,
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{
						corev1.ReadWriteOnce,
					},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: resource.MustParse("20Gi"),
						},
					},
				},
			},
		},
	}

	return obj
}

func int32p(v int32) *int32 {
	i := v
	return &i
}

func int64p(v int64) *int64 {
	i := v
	return &i
}