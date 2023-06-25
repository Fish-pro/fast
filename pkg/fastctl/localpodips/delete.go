package localpodips

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"

	bpfmap "github.com/fast-io/fast/pkg/bpf/map"
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
		Short:   "delete local pod ip eBPF map",
		Long:    "delete local pod ip eBPF map",
		Example: fmt.Sprintf("    %s localpodips --pod-ip 10.244.5.100 --ns-index 3 --ns-mac foo --host-index 3 --host-mac bar", name),
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
	localIpsMap := bpfmap.GetLocalPodIpsMap()
	podIP := util.InetIpToUInt32(o.podIP)

	if err := localIpsMap.Delete(podIP); err != nil {
		return fmt.Errorf("failed to delete pod ip %s from map: %w", o.podIP, err)
	}
	fmt.Fprintf(o.Out, "delete pod ip %s from map successfully\n", o.podIP)
	return nil
}
