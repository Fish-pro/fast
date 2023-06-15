package fastctl

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kubectl/pkg/util/templates"

	"github.com/fast-io/fast/pkg/fastctl/clusterpodips"
	"github.com/fast-io/fast/pkg/fastctl/localdev"
	"github.com/fast-io/fast/pkg/fastctl/localpodips"
)

func NewFastCtlCommand(rootCmd string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   rootCmd,
		Short: fmt.Sprintf("%s is a ctl for Fast", rootCmd),
		Long:  fmt.Sprintf("%s is a ctl for Fast", rootCmd),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	ioStreams := genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}
	groups := templates.CommandGroups{
		templates.CommandGroup{
			Message: "Basic Command",
			Commands: []*cobra.Command{
				clusterpodips.NewClusterPodIpsCommand(rootCmd, ioStreams),
				localpodips.NewLocalPodIpsCommand(rootCmd, ioStreams),
				localdev.NewLocalDevCommand(rootCmd, ioStreams),
			},
		},
	}
	groups.Add(cmd)
	templates.ActsAsRootCommand(cmd, []string{"options"}, groups...)

	return cmd
}
