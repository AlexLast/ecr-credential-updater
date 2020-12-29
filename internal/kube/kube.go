package kube

import (
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// BuildClient returns a new kubernetes client, the
// client will use default kubeconfig configuration if present
// otherwise it will use in cluster configuration
func BuildClient() (*kubernetes.Clientset, error) {
	var config *rest.Config

	// Default .kube/config path
	kubeConfigPath := filepath.Join(homedir.HomeDir(), ".kube", "config")

	// Check if kubeconfig exists
	_, err := os.Stat(kubeConfigPath)

	if err == nil {
		// If kubeconfig exists in our home directory use that to authenticate
		config, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)

		if err != nil {
			return nil, err
		}

	} else {
		// Otherwise use in-cluster config
		config, err = rest.InClusterConfig()

		if err != nil {
			return nil, err
		}
	}

	// Build client with kubeconfig file or in-cluster config
	return kubernetes.NewForConfig(config)
}
