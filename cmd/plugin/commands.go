package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/dodopizza/stand-schedule-policy-controller/internal/plugin"
)

// NewVersionCommand return command that returns plugin version
func NewVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Args:  cobra.NoArgs,
		Short: "Print your cli version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(plugin.GetVersion())
		},
	}
}

// NewStartupCommand return command that starts stand
func NewStartupCommand() *cobra.Command {
	h := plugin.NewStartupHandler()
	cmd := &cobra.Command{
		Use:   h.String(),
		Args:  cobra.ExactArgs(1),
		Short: "Startup stand",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := h.Setup(args[0]); err != nil {
				return err
			}
			return h.Run()
		},
	}
	cmd.Flags().AddFlagSet(h.SetupFlags())

	return cmd
}

// NewShutdownCommand return command that stops stand
func NewShutdownCommand() *cobra.Command {
	h := plugin.NewShutdownHandler()
	cmd := &cobra.Command{
		Use:   h.String(),
		Args:  cobra.ExactArgs(1),
		Short: "Shutdown stand",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := h.Setup(args[0]); err != nil {
				return err
			}
			return h.Run()
		},
	}
	cmd.Flags().AddFlagSet(h.SetupFlags())

	return cmd
}
