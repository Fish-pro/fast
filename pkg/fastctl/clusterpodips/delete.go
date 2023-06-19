package clusterpodips

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"

	bpfmap "github.com/fast-io/fast/bpf/map"
	"github.com/fast-io/fast/pkg/util"
)

type deleteOptions struct {
	genericclioptions.IOStreams

	podIP string
}

func newDeleteOptions(ioStream genericclioptions.IOStreams) *deleteOptions {
	return &deleteOptions{
		IOStreams: ioStream,
	}
}

func NewDeleteCommand(name string, ioStreaam genericclioptions.IOStreams) *cobra.Command {
	o := newDeleteOptions(ioStreaam)
	cmd := &cobra.Command{
		Use:     "delete",
		Aliases: []string{"del"},
		Short:   "delete cluster pod ip eBPF map",
		Long:    "delete cluster pod ip eBPF map",
		Example: fmt.Sprintf("    %s clusterpodips delete --pod-ip 10.244.0.1 --node-ip 10.29.15.48r", name),
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(cmd, args))
			cmdutil.CheckErr(o.Validate(args))
			cmdutil.CheckErr(o.Run())
		},
	}
	cmd.Flags().StringVarP(&o.podIP, "pod-ip", "", o.podIP, "The pod-ip define the eBPF map key")

	return cmd
}

func (o *deleteOptions) Complete(cmd *cobra.Command, args []string) error {
	return nil
}

func (o *deleteOptions) Validate(args []string) error {
	if len(o.podIP) == 0 {
		return fmt.Errorf("pod-ip is required")
	}
	return nil
}

func (o *deleteOptions) Run() error {
	clusterIpsMap := bpfmap.GetClusterPodIpsMap()
	podIP := util.InetIpToUInt32(o.podIP)

	if err := clusterIpsMap.Delete(podIP); err != nil {
		return fmt.Errorf("failed to delete pod ip %s from map: %w", o.podIP, err)
	}
	fmt.Fprintf(o.Out, "delete pod ip %s from map successfully\n", o.podIP)
	return nil
}
