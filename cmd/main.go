package main

import (
	"fmt"
	"k8s-token-hunter/app"
	"k8s-token-hunter/model"
	"os"

	"github.com/spf13/cobra"
)

// Define a global instance of Flags to hold flag values
var flags model.Config

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "k8s-token-hunter",
	Short: "K8s Token Hunter is a tool to hunt Kubernetes tokens",
	Long:  `K8s Token Hunter is a tool that exploits the /var/log host path mount to hunt Kubernetes tokens`,
	Run: func(cmd *cobra.Command, args []string) {
		// Validate required flags
		requiredFlags := []struct {
			value, name string
		}{
			{flags.SuffixFilePath, "suffix-file"},
			{flags.HostMountPath, "mount-path"},
		}

		for _, r := range requiredFlags {
			if r.value == "" {
				fmt.Fprintf(os.Stderr, "Error: -%s is a required flag\n", r.name)
				cmd.Help()
				os.Exit(1)
			}
		}

		app.Run(flags)
	},
}

func init() {
	rootCmd.Flags().StringVarP(&flags.ServiceAccountToken, "token", "t", "", "Service account token")
	rootCmd.Flags().StringVarP(&flags.SuffixFilePath, "suffix-file", "s", "", "Suffix file path (required)")
	rootCmd.Flags().StringVarP(&flags.HostMountPath, "mount-path", "m", "", "According to the pod manifest, the /var/log host path has been mounted to the pod (required)")
	rootCmd.Flags().StringVarP(&flags.KubernetesServiceHost, "kubernetes-service-host", "k", "", "Kubernetes API server IP address")
	rootCmd.Flags().StringVarP(&flags.TargetNamespace, "target-namespace", "n", "", "Namespace to target")
	rootCmd.Flags().StringVarP(&flags.SuffixGeneratorFile, "generate-suffix-list", "l", "", "File path for the suffix combintations list")
	rootCmd.Flags().StringVarP(&flags.OutputPath, "output-path", "o", "/tmp/k8s-token-scanner_output.txt", "Output path for the hunted tokens")

	// Mark required flags
	rootCmd.MarkFlagRequired("suffix-file")
	rootCmd.MarkFlagRequired("mount-path")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error during command execution: %v\n", err)
		os.Exit(1)
	}
}
