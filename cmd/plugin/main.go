package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/dodopizza/stand-schedule-policy-controller/internal/plugin"
)

func main() {
	cmd := &cobra.Command{
		Use:               plugin.Name,
		Short:             "Start or stop stand",
		Version:           plugin.GetVersion(),
		SilenceUsage:      true,
		DisableAutoGenTag: true,
	}

	cmd.AddCommand(NewStartupCommand())
	cmd.AddCommand(NewShutdownCommand())
	cmd.AddCommand(NewVersionCommand())

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
