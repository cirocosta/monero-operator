package reconciler

import (
	"fmt"

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

func TorHiddenServiceSecretName(nodeSet *v1alpha1.MoneroNodeSet) string {
	return nodeSet.Name + "-" + "tor"
}

func NewTorHiddenServiceSecret(nodeSet *v1alpha1.MoneroNodeSet) *corev1.Secret {
	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: corev1.SchemeGroupVersion.Identifier(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      TorHiddenServiceSecretName(nodeSet),
			Namespace: nodeSet.Namespace,
			Labels: map[string]string{
				"utxo.com.br/tor": "v3",
			},
		},
	}
}

func TorProxyConfigMapName(nodeSet *v1alpha1.MoneroNodeSet) string {
	return nodeSet.Name + "-" + "tor-proxy"
}

func NewTorProxyConfigMap(nodeSet *v1alpha1.MoneroNodeSet) *corev1.ConfigMap {
	torrc := `SOCKSPort 9050
ControlPort 9051`

	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: corev1.SchemeGroupVersion.Identifier(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      TorProxyConfigMapName(nodeSet),
			Namespace: nodeSet.Namespace,
		},
		Data: map[string]string{
			"torrc": torrc,
		},
	}
}

func TorHiddenServiceConfigMapName(nodeSet *v1alpha1.MoneroNodeSet) string {
	return nodeSet.Name + "-" + "tor-hidden-service"
}

func NewTorHiddenServiceConfigMap(nodeSet *v1alpha1.MoneroNodeSet) *corev1.ConfigMap {
	torrc := fmt.Sprintf(`HiddenServiceDir /tor
HiddenServicePort 18089 %s:18089
HiddenServicePort 18083 %s:18083
HiddenServiceVersion 3`, MoneroServiceName(nodeSet), MoneroServiceName(nodeSet))

	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: corev1.SchemeGroupVersion.Identifier(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      TorHiddenServiceConfigMapName(nodeSet),
			Namespace: nodeSet.Namespace,
		},
		Data: map[string]string{
			"torrc": torrc,
		},
	}
}

func TorHiddenServiceDeploymentName(nodeSet *v1alpha1.MoneroNodeSet) string {
	return nodeSet.Name + "-tor-proxy"
}

func NewTorHiddenServiceDeployment(nodeSet *v1alpha1.MoneroNodeSet) *appsv1.Deployment {
	o := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      TorHiddenServiceDeploymentName(nodeSet),
			Namespace: nodeSet.Namespace,
		},

		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: appsv1.SchemeGroupVersion.Identifier(),
		},

		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: AppLabel(TorHiddenServiceDeploymentName(nodeSet)),
			},
			RevisionHistoryLimit: pointer.Int32Ptr(0),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: AppLabel(TorHiddenServiceDeploymentName(nodeSet)),
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						NewTorHiddenServiceVolume(nodeSet),
					},
					Containers: []corev1.Container{
						NewTornetesContainer(nodeSet),
					},
				},
			},
		},
	}

	return o
}

func TorVolumeName(nodeSet *v1alpha1.MoneroNodeSet) string {
	return "tor"
}

func NewTorHiddenServiceVolume(nodeSet *v1alpha1.MoneroNodeSet) corev1.Volume {
	return corev1.Volume{
		Name: "tor",
		VolumeSource: corev1.VolumeSource{
			Projected: &corev1.ProjectedVolumeSource{
				Sources: []corev1.VolumeProjection{
					{
						Secret: &corev1.SecretProjection{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: TorHiddenServiceSecretName(nodeSet),
							},
						},
					},
					{
						ConfigMap: &corev1.ConfigMapProjection{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: TorHiddenServiceConfigMapName(nodeSet),
							},
						},
					},
				},
			},
		},
	}
}

func NewTorProxyVolume(nodeSet *v1alpha1.MoneroNodeSet) corev1.Volume {
	return corev1.Volume{
		Name: "tor",
		VolumeSource: corev1.VolumeSource{
			Projected: &corev1.ProjectedVolumeSource{
				Sources: []corev1.VolumeProjection{
					{
						ConfigMap: &corev1.ConfigMapProjection{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: TorProxyConfigMapName(nodeSet),
							},
						},
					},
				},
			},
		},
	}
}

func NewTornetesContainer(nodeSet *v1alpha1.MoneroNodeSet) corev1.Container {
	volumeMounts := []corev1.VolumeMount{
		{
			Name:      TorVolumeName(nodeSet),
			MountPath: "/tor-original",
		},
	}

	command := []string{
		"tornetes",
		"run",
		"--source=/tor-original/..data",
		"--destination=/tor",
	}

	return corev1.Container{
		Name:         "tornetes",
		Image:        "index.docker.io/utxobr/tornetes@sha256:3d103a73bca66fb27416b6bd23d3c66cd363157cd9f9d5a159157560ea4c48bf",
		Command:      command,
		VolumeMounts: volumeMounts,
	}
}

func NewMonerodContainer(nodeSet *v1alpha1.MoneroNodeSet) corev1.Container {
	defaultArgs := []string{
		"--data-dir=" + MonerodDataVolumeMountPath,
		"--log-file=/dev/stdout",

		"--non-interactive",
		"--no-zmq",
		"--no-igd",

		"--p2p-bind-ip=0.0.0.0",
		"--p2p-bind-port=18080",

		"--rpc-restricted-bind-ip=0.0.0.0",
		"--rpc-restricted-bind-port=18089",
	}

	if nodeSet.Spec.Tor.Enabled {
		defaultArgs = append(defaultArgs,
			"--tx-proxy=tor,127.0.0.1:9050",
			fmt.Sprintf("--anonymous-inbound=%s,%s", nodeSet.Status.Tor.Address+":18083", "127.0.0.1:18083"),
		)
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
		Resources: corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				corev1.ResourceHugePagesPrefix + "2Mi": resource.MustParse("1Gi"),
				corev1.ResourceMemory:                  resource.MustParse("1Gi"),
			},
			Requests: corev1.ResourceList{},
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

	if nodeSet.Spec.HardAntiAffinity {
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

func NewMoneroStatefulSet(nodeSet *v1alpha1.MoneroNodeSet) *appsv1.StatefulSet {
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
		ServiceName:          MoneroServiceName(nodeSet),
		Replicas:             pointer.Int32Ptr(int32(nodeSet.Spec.Replicas)),
		RevisionHistoryLimit: pointer.Int32Ptr(0),
		Selector: &metav1.LabelSelector{
			MatchLabels: AppLabel(nodeSet.Name),
		},
		Template: NewPodTemplateSpec(nodeSet),
		VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
			NewVolumeClaimTemplate(nodeSet),
		},
	}

	// tor enabled, let's include the extra volume, as well as secret
	//
	// ps.: this is _true_ only for `replicas==1`
	//
	if nodeSet.Spec.Tor.Enabled {
		obj.Spec.Template.Spec.Volumes = []corev1.Volume{
			NewTorProxyVolume(nodeSet),
		}

		obj.Spec.Template.Spec.Containers = append(obj.Spec.Template.Spec.Containers,
			NewTornetesContainer(nodeSet),
		)
	}

	return obj
}

func TorHiddenServiceServiceName(nodeSet *v1alpha1.MoneroNodeSet) string {
	return nodeSet.Name + "-tor-hidden-service"
}

func NewTorHiddenServiceService(nodeSet *v1alpha1.MoneroNodeSet) *corev1.Service {
	obj := &corev1.Service{}

	obj.TypeMeta = metav1.TypeMeta{
		Kind:       "Service",
		APIVersion: corev1.SchemeGroupVersion.Identifier(),
	}

	l := AppLabel(nodeSet.Name)

	obj.ObjectMeta = metav1.ObjectMeta{
		Name:      TorHiddenServiceServiceName(nodeSet),
		Namespace: nodeSet.Namespace,
		Labels:    l,
	}

	obj.Spec = corev1.ServiceSpec{
		Selector: l,
		Ports: []corev1.ServicePort{
			{
				Name:       TorP2PPortName,
				Port:       int32(TorP2PPortNumber),
				TargetPort: intstr.FromInt(int(TorP2PPortNumber)),
				Protocol:   corev1.ProtocolTCP,
			},
		},
	}
	return obj
}

func MoneroServiceName(nodeSet *v1alpha1.MoneroNodeSet) string {
	return nodeSet.Name
}

func NewMoneroService(nodeSet *v1alpha1.MoneroNodeSet) *corev1.Service {
	obj := &corev1.Service{}

	obj.TypeMeta = metav1.TypeMeta{
		Kind:       "Service",
		APIVersion: corev1.SchemeGroupVersion.Identifier(),
	}

	l := AppLabel(nodeSet.Name)

	obj.ObjectMeta = metav1.ObjectMeta{
		Name:      MoneroServiceName(nodeSet),
		Namespace: nodeSet.Namespace,
		Labels:    l,
	}

	obj.Spec = corev1.ServiceSpec{
		Selector: l,
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

	if nodeSet.Spec.Service.Type == "NodePort" {
		obj.Spec.Type = corev1.ServiceTypeNodePort
		for idx := range obj.Spec.Ports {
			obj.Spec.Ports[idx].NodePort = obj.Spec.Ports[idx].Port + 30000 - 18000
		}
	}

	return obj
}
