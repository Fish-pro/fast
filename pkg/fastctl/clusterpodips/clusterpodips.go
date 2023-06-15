package clusterpodips

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func NewClusterPodIpsCommand(name string, ioStreams genericclioptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "clusterpodips COMMAND",
		Aliases: []string{"cips"},
		Short:   "Manage cluster pod ips on the Fast",
		Long:    "Manage cluster pod ips on the Fast",
		Run: func(cmd *cobra.Command, args []string) {

		},
	}
	cmd.AddCommand(NewListCommand(name, ioStreams))
	return cmd
}
