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
	// 获取当前 Gateway 部署状态
	status, err := gw.appCache.GetDeployStatus()
	if err != nil {
		gw.log.Error(err)
		return err
	}

	// 如果已经部署完成，不需要重复更新
	if status == app_cache.Deployed {
		return nil
	}

	// 获取 Redis 中所有 HTTP 应用的：
	// 域名 → 后端地址
	mp, err := gw.appCache.GetAll()
	if err != nil {
		gw.log.Error(err)
		return err
	}

	// 根据应用缓存生成 Nginx 配置
	serverConf := gw.generateNginxConf(mp)
	if serverConf == "" {
		err = zerror.NewByMsg("服务配置获取失败")
		gw.log.Error(err)
		return err
	}

	// 将新的 Nginx 配置部署到 K8s Gateway
	err = gw.deployGatewayK8s(serverConf)
	if err != nil {
		gw.log.Error(err)
		return err
	}

	// Gateway 更新成功后，将状态改成“已部署”
	err = gw.appCache.SetDeployStatus(app_cache.Deployed)
	if err != nil {
		gw.log.Error(err)
		return err
	}

	return nil
}

func (gw *Gateway) deployGatewayK8s(configData string) error {
	// 创建 Kubernetes 操作上下文
	ctx := context.Background()

	// 获取 Kubernetes ClientSet，用于操作 ConfigMap、DaemonSet、Service 等资源
	cli := k8s.GetK8sClientSet()

	// 获取 Gateway 的相关配置
	serviceName := config.GetConfig().NginxGateway.AppName
	namespaceName := config.GetConfig().NginxGateway.Namespace
	image := config.GetConfig().NginxGateway.Image
	confPath := config.GetConfig().NginxGateway.ConfigPath

	// =========================
	// 1. 创建 / 更新 ConfigMap
	// =========================

	// ConfigMap 名称，例如：tunnel-gateway-conf
	configName := fmt.Sprintf("%s-conf", serviceName)

	// Nginx 配置文件名称，例如：tunnel-gateway.conf
	nginxConfigName := fmt.Sprintf("%s.conf", serviceName)

	// 标记 Nginx 配置是否发生变化
	// 如果配置没有变化，则后面的 DaemonSet 和 Service 不需要更新
	configChange := true

	// 查询指定名称的 ConfigMap 是否已经存在
	confList, err := cli.CoreV1().ConfigMaps(namespaceName).List(
		ctx,
		metav1.ListOptions{
			FieldSelector: fmt.Sprintf("metadata.name=%s", configName),
		},
	)
	if err != nil {
		gw.log.Error(err)
		return err
	}

	// 创建 ConfigMap 对象
	// Data 中保存真正的 Nginx 配置文件内容
	cmSpec := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: configName,
		},
		Data: map[string]string{
			nginxConfigName: string(configData),
		},
	}

	// ConfigMap 不存在，直接创建
	if len(confList.Items) == 0 {
		_, err = cli.CoreV1().ConfigMaps(namespaceName).Create(
			ctx,
			cmSpec,
			metav1.CreateOptions{},
		)
		if err != nil {
			gw.log.Error(err)
			return err
		}
	} else {
		// ConfigMap 已经存在

		// 比较旧的 Nginx 配置和新的配置
		if confList.Items[0].Data[nginxConfigName] == configData {
			// 配置没有发生变化
			configChange = false
		} else {
			// 配置发生变化，需要更新 ConfigMap

			// Kubernetes 更新资源时需要携带当前 ResourceVersion
			cmSpec.ObjectMeta.ResourceVersion = confList.Items[0].ResourceVersion

			_, err = cli.CoreV1().ConfigMaps(namespaceName).Update(
				ctx,
				cmSpec,
				metav1.UpdateOptions{},
			)
			if err != nil {
				gw.log.Error(err)
				return err
			}
		}
	}

	// =========================
	// 2. 创建 / 更新 DaemonSet
	// =========================

	// Gateway 对应的 DaemonSet 名称
	dsName := fmt.Sprintf("%s-ds", serviceName)

	// 查询 DaemonSet 是否已经存在
	dsList, err := cli.AppsV1().DaemonSets(namespaceName).List(
		ctx,
		metav1.ListOptions{
			FieldSelector: fmt.Sprintf("metadata.name=%s", dsName),
		},
	)
	if err != nil {
		gw.log.Error(err)
		return err
	}

	// Gateway Pod 暴露 80 端口
	podPorts := []corev1.ContainerPort{
		{
			ContainerPort: 80,
			Protocol:      "TCP",
		},
	}

	// Gateway Service 的端口配置
	svcPorts := []corev1.ServicePort{
		{
			// Service 端口名称
			Name: fmt.Sprintf(
				"tcp-%d",
				config.GetConfig().NginxGateway.Port,
			),

			Protocol: "TCP",

			// Service 流量转发到 Pod 的 80 端口
			TargetPort: intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: 80,
			},

			// Service 对外暴露的端口
			Port: int32(config.GetConfig().NginxGateway.Port),

			// NodePort 端口
			NodePort: int32(config.GetConfig().NginxGateway.Port),
		},
	}

	// 创建 Gateway DaemonSet
	dsSpec := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: dsName,
		},

		Spec: appsv1.DaemonSetSpec{

			// DaemonSet 根据这个 Label 找到自己管理的 Pod
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"service-name": serviceName,
				},
			},

			// Pod 模板
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					// 给 Gateway Pod 添加 Label
					Labels: map[string]string{
						"service-name": serviceName,
					},
				},

				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							// 容器名称
							Name: serviceName,

							// Gateway 镜像
							Image: image,

							// 容器监听 80 端口
							Ports: podPorts,

							// 挂载 ConfigMap
							VolumeMounts: []corev1.VolumeMount{
								{
									Name: configName,

									// ConfigMap 中的配置文件
									// 挂载到 Nginx 的配置目录
									MountPath: fmt.Sprintf(
										"%s/%s",
										confPath,
										nginxConfigName,
									),

									// 只挂载指定的配置文件
									SubPath: nginxConfigName,
								},
							},
						},
					},

					// 定义 Pod 使用的 Volume
					Volumes: []corev1.Volume{
						{
							Name: configName,

							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{

									// 指定使用哪个 ConfigMap
									LocalObjectReference: corev1.LocalObjectReference{
										Name: configName,
									},

									// 指定挂载 ConfigMap 中的哪个文件
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

	// DaemonSet 不存在，创建 DaemonSet
	if len(dsList.Items) == 0 {
		_, err = cli.AppsV1().DaemonSets(namespaceName).Create(
			ctx,
			dsSpec,
			metav1.CreateOptions{},
		)
		if err != nil {
			gw.log.Error(err)
			return err
		}
	} else {
		// DaemonSet 已经存在

		// 只有 Nginx 配置发生变化时才更新 DaemonSet
		if configChange {

			// 更新 K8s 资源需要携带 ResourceVersion
			dsSpec.ResourceVersion = dsList.Items[0].ResourceVersion

			_, err = cli.AppsV1().DaemonSets(namespaceName).Update(
				ctx,
				dsSpec,
				metav1.UpdateOptions{},
			)
			if err != nil {
				gw.log.Error(err)
				return err
			}
		}
	}

	// =========================
	// 3. 创建 / 更新 Service
	// =========================

	// Gateway Service 名称
	svcName := fmt.Sprintf("%s-svc", serviceName)

	// 创建 Service 对象
	svcSpec := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: svcName,
		},

		Spec: corev1.ServiceSpec{

			// Service 将流量转发给带有这个 Label 的 Gateway Pod
			Selector: map[string]string{
				"service-name": serviceName,
			},

			// 使用 NodePort 暴露 Gateway
			Type: corev1.ServiceTypeNodePort,

			// Service 端口配置
			Ports: svcPorts,
		},
	}

	// 查询 Service 是否已经存在
	svcList, err := cli.CoreV1().Services(namespaceName).List(
		ctx,
		metav1.ListOptions{
			FieldSelector: fmt.Sprintf("metadata.name=%s", svcName),
		},
	)
	if err != nil {
		gw.log.Error(err)
		return err
	}

	// Service 不存在，创建 Service
	if len(svcList.Items) == 0 {
		_, err = cli.CoreV1().Services(namespaceName).Create(
			ctx,
			svcSpec,
			metav1.CreateOptions{},
		)
		if err != nil {
			gw.log.Error(err)
			return err
		}
	} else {
		// Service 已经存在
		// 只有配置发生变化时才更新 Service
		if configChange {

			// 更新 K8s 资源需要携带 ResourceVersion
			svcSpec.ResourceVersion = svcList.Items[0].ResourceVersion

			_, err = cli.CoreV1().Services(namespaceName).Update(
				ctx,
				svcSpec,
				metav1.UpdateOptions{},
			)
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
	// 保存最终生成的 Nginx 配置
	confStr := ""

	// 遍历所有应用
	// k = 域名
	// v = 后端地址
	for k, v := range mp {

		// 根据 Nginx 配置模板生成一个 server 配置
		// k 和 v 分别填入域名和后端地址
		confStr += fmt.Sprintf(serverConfTemplate, k, v)

		// 每个 server 配置之间换一行
		confStr += "\n"
	}

	// 返回完整的 Nginx 配置
	return confStr
}
