package localpodips

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func NewLocalPodIpsCommand(name string, ioStreams genericclioptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "localpodips COMMAND",
		Aliases: []string{"lips"},
		Short:   "Manage local pod ips on the Fast",
		Long:    "Manage local pod ips on the Fast",
		Run: func(cmd *cobra.Command, args []string) {

		},
	}
	cmd.AddCommand(NewListCommand(name, ioStreams))
	return cmd
}
