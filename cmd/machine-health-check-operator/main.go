package main

import (
	"flag"
	"os"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
)

var componentNamespace = "openshift-machine-api"

const (
	componentName = "machine-health-check-operator"
)

var (
	rootCmd = &cobra.Command{
		Use:   componentName,
		Short: "Run machine health check operator that will deploy machine health check controller",
		Long:  "",
	}
	config string
)

func init() {
	rootCmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)
}

func main() {
	if namespace, ok := os.LookupEnv("COMPONENT_NAMESPACE"); ok {
		componentNamespace = namespace
	}
	if err := rootCmd.Execute(); err != nil {
		glog.Exitf("Error executing machine health check operator: %v", err)
	}
}
