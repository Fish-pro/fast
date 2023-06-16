package localdev

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

func NewLocalDevCommand(name string, ioStreams genericclioptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "localdev COMMAND",
		Aliases: []string{"ldev"},
		Short:   "Manage local device on the Fast",
		Long:    "Manage local device on the Fast",
		Run:     cmdutil.DefaultSubCommandRun(ioStreams.ErrOut),
	}
	cmd.AddCommand(NewListCommand(name, ioStreams))
	return cmd
}
