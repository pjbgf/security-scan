package helper

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"k8s.io/client-go/discovery"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func ServerVersion() (string, error) {
	configPath := os.Getenv("KUBECONFIG")
	if configPath == "" {
		if home := homedir.HomeDir(); home != "" {
			configPath = filepath.Join(home, ".kube", "config")
		} else {
			return "", fmt.Errorf("KUBECONFIG environment variable not set and home directory not found to locate default kubeconfig.")
		}
	}

	slog.Debug("Attempting to consume kubeconfig", "path", configPath)
	config, err := clientcmd.BuildConfigFromFlags("", configPath)
	if err != nil {
		return "", fmt.Errorf("error building kubeconfig: %v", err)
	}

	client, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return "", fmt.Errorf("cannot create client: %v", err)
	}

	version, err := client.ServerVersion()
	if err != nil {
		return "", fmt.Errorf("cannot get server version: %v", err)
	}
	return version.String(), nil
}
