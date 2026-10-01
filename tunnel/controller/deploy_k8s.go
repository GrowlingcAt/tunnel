package controller

import (
	"context"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"strings"
	"tunnel/pkg/config"
	"tunnel/pkg/k8s"
)

func (c *App) deployK8s(serverID int64, configData []byte, ports map[uint32]string) (string, error) {
	ctx := context.Background()
	cli := k8s.GetK8sClientSet()
	serviceName := c.getServiceName(serverID)
	namespaceName := config.GetConfig().TunnelServer.Namespace
	replicas := int32(config.GetConfig().TunnelServer.Replicas)
	image := config.GetConfig().TunnelServer.Image
	// 配置文件创建
	configName := serviceName + "-conf"
	confList, err := cli.CoreV1().ConfigMaps(namespaceName).List(ctx, metav1.ListOptions{FieldSelector: "metadata.name=" + configName})
	if err != nil {
		c.log.Error(err)
		return "", err
	}
	cmSpec := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: configName,
		},
		Data: map[string]string{"config.yaml": string(configData)},
	}
	if len(confList.Items) == 0 {
		_, err = cli.CoreV1().ConfigMaps(namespaceName).Create(ctx, cmSpec, metav1.CreateOptions{})
		if err != nil {
			c.log.Error(err)
			return "", err
		}
	} else {
		cmSpec.ResourceVersion = confList.Items[0].ResourceVersion
		_, err = cli.CoreV1().ConfigMaps(namespaceName).Update(ctx, cmSpec, metav1.UpdateOptions{})
		if err != nil {
			c.log.Error(err)
			return "", err
		}
	}
	// Deployment创建
	deployName := serviceName + "-deploy"
	deployList, err := cli.AppsV1().Deployments(namespaceName).List(ctx, metav1.ListOptions{FieldSelector: "metadata.name=" + deployName})
	if err != nil {
		c.log.Error(err)
		return "", err
	}
	podPorts := make([]corev1.ContainerPort, 0, len(ports))
	svcPorts := make([]corev1.ServicePort, 0, len(ports))
	for port, protocol := range ports {
		p := corev1.ContainerPort{
			ContainerPort: int32(port),
			Protocol:      corev1.Protocol(strings.ToUpper(protocol)),
		}
		sp := corev1.ServicePort{
			Name:     fmt.Sprintf("%s-%d", strings.ToLower(protocol), port),
			Protocol: p.Protocol,
			Port:     int32(port),
			TargetPort: intstr.IntOrString{
				IntVal: p.ContainerPort,
				Type:   intstr.Int,
			},
			NodePort: int32(port),
		}
		podPorts = append(podPorts, p)
		svcPorts = append(svcPorts, sp)
	}
	deploySpec := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: deployName,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"service-name": serviceName,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"service-name": serviceName,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  config.GetConfig().TunnelServer.AppName,
							Image: image,
							Ports: podPorts,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "tunnel-server-conf",
									MountPath: "/app/config.yaml",
									SubPath:   "config.yaml",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "tunnel-server-conf",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: configName,
									},
									Items: []corev1.KeyToPath{
										{
											Key:  "config.yaml",
											Path: "config.yaml",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	if len(deployList.Items) == 0 {
		_, err = cli.AppsV1().Deployments(namespaceName).Create(ctx, deploySpec, metav1.CreateOptions{})
		if err != nil {
			c.log.Error(err)
			return "", err
		}
	} else {
		deploySpec.ResourceVersion = deployList.Items[0].ResourceVersion
		_, err = cli.AppsV1().Deployments(namespaceName).Update(ctx, deploySpec, metav1.UpdateOptions{})
		if err != nil {
			c.log.Error(err)
			return "", err
		}
	}

	// Service创建
	svcName := serviceName + "-svc"
	fmt.Println(svcName)
	svcList, err := cli.CoreV1().Services(namespaceName).List(ctx, metav1.ListOptions{FieldSelector: "metadata.name=" + svcName})
	if err != nil {
		c.log.Error(err)
		return "", err
	}
	svcSpec := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: svcName,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"service-name": serviceName,
			},
			Type:  corev1.ServiceTypeNodePort,
			Ports: svcPorts,
		},
	}
	if len(svcList.Items) == 0 {
		_, err = cli.CoreV1().Services(namespaceName).Create(ctx, svcSpec, metav1.CreateOptions{})
		if err != nil {
			c.log.Error(err)
			return "", err
		}
	} else {
		svcSpec.ResourceVersion = svcList.Items[0].ResourceVersion
		_, err = cli.CoreV1().Services(namespaceName).Update(ctx, svcSpec, metav1.UpdateOptions{})
		if err != nil {
			c.log.Error(err)
			return "", err
		}
	}
	return serviceName, nil
}
