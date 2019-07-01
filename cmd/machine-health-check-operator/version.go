package main

import (
	"flag"
	"fmt"

	"github.com/openshift/machine-health-check-operator/pkg/version"
	"github.com/spf13/cobra"
)

var (
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version number of Machine Health Check Operator",
		Long:  `All software has versions. This is Machine Health Check Operator.`,
		Run:   runVersionCmd,
	}
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

func runVersionCmd(cmd *cobra.Command, args []string) {
	flag.Set("logtostderr", "true")
	flag.Parse()

	program := "Machine Health Check Operator"

	fmt.Println(program, version.Get().String())
}
