package localdev

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func NewLocalDevCommand(name string, ioStreams genericclioptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "localdev COMMAND",
		Aliases: []string{"ldev"},
		Short:   "Manage local device on the Fast",
		Long:    "Manage local device on the Fast",
		Run: func(cmd *cobra.Command, args []string) {

		},
	}
	cmd.AddCommand(NewListCommand(name, ioStreams))
	return cmd
}
