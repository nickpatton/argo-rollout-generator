package main

import (
	"fmt"

	rolloutsv1a1 "github.com/argoproj/argo-rollouts/pkg/apis/rollouts/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func main() {

	// fields that are set in the deployment pipeline
	var replicas int32 = 3
	var defaultRevisionHistory int32 = 0
	helperTrue := true
	var scaleDownDelaySeconds int32 = 60
	var applicationName string = "dancing-bears"
	var image string = "ghcr.io/nickpatton/some-dancing-bears:v0.0.4"
	var namespace string = "bear-system"
	var appPort int = 5000
	labels := map[string]string{
		"bear-type": "polar-bear",
	}

	blueService := corev1.Service{
		TypeMeta: v1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      applicationName + "-svc",
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Name:       "app",
					Protocol:   corev1.ProtocolTCP,
					Port:       443,
					TargetPort: intstr.FromInt(int(appPort)),
				},
			},
			Selector: labels,
		},
	}

	greenService := corev1.Service{
		TypeMeta: v1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      applicationName + "-preview-svc",
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Name:       "app",
					Protocol:   corev1.ProtocolTCP,
					Port:       443,
					TargetPort: intstr.FromInt(int(appPort)),
				},
			},
			Selector: labels,
		},
	}

	// native rollout object is generated using values set in pipeline
	rollout := rolloutsv1a1.Rollout{
		TypeMeta: v1.TypeMeta{
			Kind:       "Rollout",
			APIVersion: "argoproj.io/v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      applicationName,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: rolloutsv1a1.RolloutSpec{
			Replicas:             &replicas,
			RevisionHistoryLimit: &defaultRevisionHistory,
			Selector: &v1.LabelSelector{
				MatchLabels: labels,
			},
			Strategy: rolloutsv1a1.RolloutStrategy{
				BlueGreen: &rolloutsv1a1.BlueGreenStrategy{
					ActiveService:         applicationName + "-svc",
					PreviewService:        applicationName + "-preview-svc",
					AutoPromotionEnabled:  &helperTrue,
					ScaleDownDelaySeconds: &scaleDownDelaySeconds,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: v1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image:           image,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Name:            applicationName,
							Ports: []corev1.ContainerPort{
								{
									Name:          "app",
									ContainerPort: 5000,
								},
							},
						},
					},
					RestartPolicy: "Always",
				},
			},
		},
	}

	yamlString := joinKubernetesObjects(blueService, greenService, rollout)
	err := writeFile(yamlString)
	if err != nil {
		fmt.Printf("Error writing file: %v", err)
	}
}
