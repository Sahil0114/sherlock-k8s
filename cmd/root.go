package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	namespace  string
	kubeconfig string
)

var rootCmd = &cobra.Command{
	Use:   "kubectl-sherlock",
	Short: "Instant Root Cause Analysis context for Kubernetes",
	Long:  "Kube-Sherlock aggregates Node conditions, multi-container logs, and events into a single, chronologically sorted timeline.",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	defaultKubeconfig := ""
	if home := homedir.HomeDir(); home != "" {
		defaultKubeconfig = filepath.Join(home, ".kube", "config")
	}
	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "Kubernetes namespace (defaults to current context namespace or 'default')")
	rootCmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", defaultKubeconfig, "Path to kubeconfig file")
}

func buildClient() (*kubernetes.Clientset, string, error) {
	// Try in-cluster first
	config, err := rest.InClusterConfig()
	if err != nil {
		// Fallback to kubeconfig
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, "", fmt.Errorf("failed to build kubernetes config: %w", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	ns := resolveNamespace()
	return clientset, ns, nil
}

func resolveNamespace() string {
	if namespace != "" {
		return namespace
	}

	// Try to get namespace from current context
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if kubeconfig != "" {
		loadingRules.ExplicitPath = kubeconfig
	}
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	ns, _, err := kubeConfig.Namespace()
	if err != nil || ns == "" {
		return "default"
	}
	return ns
}

func exitWithError(msg string) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", msg)
	os.Exit(1)
}
