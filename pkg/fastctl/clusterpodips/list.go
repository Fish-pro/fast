package clusterpodips

import (
	"fmt"

	"github.com/cilium/ebpf"
	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"

	bpfmap "github.com/fast-io/fast/pkg/bpf/map"
	"github.com/fast-io/fast/pkg/util"
)

type listOptions struct {
	genericclioptions.IOStreams

	clusterIpsMap *ebpf.Map
}

func newListOptions(ioStream genericclioptions.IOStreams) *listOptions {
	return &listOptions{
		IOStreams:     ioStream,
		clusterIpsMap: bpfmap.GetClusterPodIpsMap(),
	}
}

func NewListCommand(name string, ioStreaam genericclioptions.IOStreams) *cobra.Command {
	o := newListOptions(ioStreaam)
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "list cluster pod ips",
		Long:    "list cluster pod ips",
		Example: fmt.Sprintf("    %s clusterips ls", name),
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
	if o.clusterIpsMap == nil {
		return fmt.Errorf("failed to load eBPF map")
	}
	return nil
}

func (o *listOptions) Run() error {
	var (
		key   bpfmap.ClusterIpsMapKey
		value bpfmap.ClusterIpsMapInfo
	)

	table := uitable.New()
	table.MaxColWidth = 80
	table.AddRow("CLUSTERIP", "VALUE")
	iter := o.clusterIpsMap.Iterate()
	for iter.Next(&key, &value) {
		table.AddRow(util.InetUint32ToIp(key.IP), util.InetUint32ToIp(value.IP))
	}
	fmt.Println(table)
	return nil
}
