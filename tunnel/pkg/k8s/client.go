package k8s

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"tunnel/pkg/log"
)

var clientSet kubernetes.Interface

func InitK8sClient(kubeConfig string) {
	var config *rest.Config
	var err error
	if kubeConfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeConfig)
	} else {
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		log.Fatal(err)
	}
	clientSet, err = kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}
}
func GetK8sClientSet() kubernetes.Interface {
	return clientSet
}
