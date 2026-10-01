package nginx_gateway

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	app_cache "tunnel/app-cache"
	"tunnel/data"
	"tunnel/pkg/config"
	redis2 "tunnel/pkg/db/redis"
	"tunnel/pkg/k8s"
	"tunnel/pkg/log"
	"tunnel/pkg/zerror"
)

type Gateway struct {
	config   *config.Config
	log      log.ILogger
	data     data.IData
	appCache *app_cache.AppCache
	redisCli redis.UniversalClient
}

func NewGateway(config *config.Config, log log.ILogger, data data.IData) *Gateway {
	appCache := app_cache.NewAppCache(config, log, data)
	redisCli := redis2.Get()
	return &Gateway{
		config:   config,
		log:      log,
		data:     data,
		appCache: appCache,
		redisCli: redisCli,
	}
}

func (gw *Gateway) DeployK8s() error {
	status, err := gw.appCache.GetDeployStatus()
	if err != nil {
		gw.log.Error(err)
		return err
	}
	if status == app_cache.Deployed {
		return nil
	}

	mp, err := gw.appCache.GetAll()
	if err != nil {
		gw.log.Error(err)
		return err
	}
	serverConf := gw.generateNginxConf(mp)
	if serverConf == "" {
		err = zerror.NewByMsg("服务配置获取失败")
		gw.log.Error(err)
		return err
	}

	err = gw.deployGatewayK8s(serverConf)
	if err != nil {
		gw.log.Error(err)
		return err
	}

	err = gw.appCache.SetDeployStatus(app_cache.Deployed)
	if err != nil {
		gw.log.Error(err)
		return err
	}

	return nil
}

func (gw *Gateway) deployGatewayK8s(configData string) error {
	ctx := context.Background()
	cli := k8s.GetK8sClientSet()
	serviceName := config.GetConfig().NginxGateway.AppName
	namespaceName := config.GetConfig().NginxGateway.Namespace
	image := config.GetConfig().NginxGateway.Image
	confPath := config.GetConfig().NginxGateway.ConfigPath

	// 配置文件创建
	configName := fmt.Sprintf("%s-conf", serviceName)
	nginxConfigName := fmt.Sprintf("%s.conf", serviceName)
	configChange := true
	confList, err := cli.CoreV1().ConfigMaps(namespaceName).List(ctx, metav1.ListOptions{FieldSelector: fmt.Sprintf("metadata.name=%s", configName)})
	if err != nil {
		gw.log.Error(err)
		return err
	}
	cmSpec := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: configName,
		},
		Data: map[string]string{nginxConfigName: string(configData)},
	}
	if len(confList.Items) == 0 {
		_, err = cli.CoreV1().ConfigMaps(namespaceName).Create(ctx, cmSpec, metav1.CreateOptions{})
		if err != nil {
			gw.log.Error(err)
			return err
		}
	} else {
		// 配置文件未发生修改
		if confList.Items[0].Data[nginxConfigName] == configData {
			configChange = false
		} else {
			cmSpec.ObjectMeta.ResourceVersion = confList.Items[0].ResourceVersion
			_, err = cli.CoreV1().ConfigMaps(namespaceName).Update(ctx, cmSpec, metav1.UpdateOptions{})
			if err != nil {
				gw.log.Error(err)
				return err
			}
		}
	}
	// DaemonSet创建
	dsName := fmt.Sprintf("%s-ds", serviceName)
	dsList, err := cli.AppsV1().DaemonSets(namespaceName).List(ctx, metav1.ListOptions{FieldSelector: fmt.Sprintf("metadata.name=%s", dsName)})
	if err != nil {
		gw.log.Error(err)
		return err
	}
	podPorts := []corev1.ContainerPort{
		{
			ContainerPort: 80,
			Protocol:      "TCP",
		},
	}
	svcPorts := []corev1.ServicePort{
		{
			Name:     fmt.Sprintf("tcp-%d", config.GetConfig().NginxGateway.Port),
			Protocol: "TCP",
			TargetPort: intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: 80,
			},
			Port:     int32(config.GetConfig().NginxGateway.Port),
			NodePort: int32(config.GetConfig().NginxGateway.Port),
		},
	}

	dsSpec := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: dsName,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"service-name": serviceName},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"service-name": serviceName},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  serviceName,
							Image: image,
							Ports: podPorts,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      configName,
									MountPath: fmt.Sprintf("%s/%s", confPath, nginxConfigName),
									SubPath:   nginxConfigName,
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: configName,
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: configName,
									},
									Items: []corev1.KeyToPath{
										{
											Key:  nginxConfigName,
											Path: nginxConfigName,
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
	if len(dsList.Items) == 0 {
		_, err = cli.AppsV1().DaemonSets(namespaceName).Create(ctx, dsSpec, metav1.CreateOptions{})
		if err != nil {
			gw.log.Error(err)
			return err
		}
	} else {
		if configChange {
			dsSpec.ResourceVersion = dsList.Items[0].ResourceVersion
			_, err = cli.AppsV1().DaemonSets(namespaceName).Update(ctx, dsSpec, metav1.UpdateOptions{})
			if err != nil {
				gw.log.Error(err)
				return err
			}
		}
	}
	// 创建service
	svcName := fmt.Sprintf("%s-svc", serviceName)
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
	svcList, err := cli.CoreV1().Services(namespaceName).List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", svcName),
	})
	if err != nil {
		gw.log.Error(err)
		return err
	}
	if len(svcList.Items) == 0 {
		_, err = cli.CoreV1().Services(namespaceName).Create(ctx, svcSpec, metav1.CreateOptions{})
		if err != nil {
			gw.log.Error(err)
			return err
		}
	} else {
		if configChange {
			svcSpec.ResourceVersion = svcList.Items[0].ResourceVersion
			_, err = cli.CoreV1().Services(namespaceName).Update(ctx, svcSpec, metav1.UpdateOptions{})
			if err != nil {
				gw.log.Error(err)
				return err
			}
		}
	}
	return nil
}

var serverConfTemplate = `server {
        listen       80;
        server_name  %s;
        location / {
            proxy_pass %s;
        }
}`

func (gw *Gateway) generateNginxConf(mp map[string]string) string {
	confStr := ""
	for k, v := range mp {
		confStr += fmt.Sprintf(serverConfTemplate, k, v)
		confStr += "\n"
	}
	return confStr
}
