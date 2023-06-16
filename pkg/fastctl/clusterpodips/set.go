package clusterpodips

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"

	bpfmap "github.com/fast-io/fast/bpf/map"
	"github.com/fast-io/fast/pkg/util"
)

type setOptions struct {
	genericclioptions.IOStreams

	podIP  string
	nodeIP string
}

func newSetOptions(ioStream genericclioptions.IOStreams) *setOptions {
	return &setOptions{
		IOStreams: ioStream,
	}
}

func NewSetCommand(name string, ioStreaam genericclioptions.IOStreams) *cobra.Command {
	o := newSetOptions(ioStreaam)
	cmd := &cobra.Command{
		Use:     "set",
		Aliases: []string{},
		Short:   "set cluster pod ip eBPF map",
		Long:    "set cluster pod ip eBPF map",
		Example: fmt.Sprintf("    %s clusterpodips set --pod-ip 10.244.0.1 --node-ip 10.29.15.48r", name),
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(cmd, args))
			cmdutil.CheckErr(o.Validate(args))
			cmdutil.CheckErr(o.Run())
		},
	}
	cmd.Flags().StringVarP(&o.podIP, "pod-ip", "", o.podIP, "The pod-ip define the eBPF map key")
	cmd.Flags().StringVarP(&o.nodeIP, "node-ip", "", o.nodeIP, "The node-ip define the node ip address")

	return cmd
}

func (o *setOptions) Complete(cmd *cobra.Command, args []string) error {
	return nil
}

func (o *setOptions) Validate(args []string) error {
	if len(o.podIP) == 0 {
		return fmt.Errorf("pod-ip is required")
	}
	if len(o.nodeIP) == 0 {
		return fmt.Errorf("node-ip is required")
	}
	return nil
}

func (o *setOptions) Run() error {
	clusterIpsMap := bpfmap.GetClusterPodIpsMap()
	podIP := util.InetIpToUInt32(o.podIP)
	nodeIP := util.InetIpToUInt32(o.nodeIP)

	if err := clusterIpsMap.Put(
		bpfmap.ClusterIpsMapKey{IP: podIP},
		bpfmap.ClusterIpsMapInfo{IP: nodeIP},
	); err != nil {
		return fmt.Errorf("failed to set cluster pod ip %s to map: %w", o.podIP, err)
	}
	fmt.Fprintf(o.Out, "set map successfully %s\n", o.podIP)
	return nil
}
