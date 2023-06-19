package localpodips

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

func NewLocalPodIpsCommand(name string, ioStreams genericclioptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "localpodips COMMAND",
		Aliases: []string{"lips"},
		Short:   "Manage local pod ips on the Fast",
		Long:    "Manage local pod ips on the Fast",
		Run:     cmdutil.DefaultSubCommandRun(ioStreams.ErrOut),
	}
	cmd.AddCommand(NewListCommand(name, ioStreams))
	cmd.AddCommand(NewSetCommand(name, ioStreams))
	cmd.AddCommand(NewDeleteCommand(name, ioStreams))
	return cmd
}
