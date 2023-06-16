package localdev

import (
	"fmt"

	"github.com/cilium/ebpf"
	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"

	bpfmap "github.com/fast-io/fast/bpf/map"
)

type listOptions struct {
	genericclioptions.IOStreams

	localDevMap *ebpf.Map
}

func newListOptions(ioStream genericclioptions.IOStreams) *listOptions {
	return &listOptions{
		IOStreams:   ioStream,
		localDevMap: bpfmap.GetLocalDevMap(),
	}
}

func NewListCommand(name string, ioStreaam genericclioptions.IOStreams) *cobra.Command {
	o := newListOptions(ioStreaam)
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "list local dev",
		Long:    "list local dev",
		Example: fmt.Sprintf("    %s localdev ls", name),
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(cmd, args))
			cmdutil.CheckErr(o.Validate(args))
			cmdutil.CheckErr(o.Run())
		},
	}
	return cmd
}

func (o *listOptions) Complete(cmd *cobra.Command, args []string) error {
	return nil
}

func (o *listOptions) Validate(args []string) error {
	return nil
}

func (o *listOptions) Run() error {
	var (
		key   bpfmap.LocalDevMapKey
		value bpfmap.LocalDevMapValue
	)

	table := uitable.New()
	table.MaxColWidth = 80
	table.AddRow("TYPE", "NAME")
	iter := o.localDevMap.Iterate()
	for iter.Next(&key, &value) {
		table.AddRow(key.Type, value.IfIndex)
	}
	fmt.Println(table)
	return nil
}
