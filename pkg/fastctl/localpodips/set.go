package localpodips

import (
	"fmt"

	"github.com/cilium/ebpf"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"

	bpfmap "github.com/fast-io/fast/pkg/bpf/map"
	"github.com/fast-io/fast/pkg/util"
)

type setOptions struct {
	genericclioptions.IOStreams

	localIpsMap *ebpf.Map

	podIP     string
	nsIndex   int
	nsMac     string
	hostIndex int
	hostMac   string
}

func newSetOptions(ioStream genericclioptions.IOStreams) *setOptions {
	return &setOptions{
		IOStreams:   ioStream,
		localIpsMap: bpfmap.GetLocalPodIpsMap(),
	}
}

func NewSetCommand(name string, ioStreaam genericclioptions.IOStreams) *cobra.Command {
	o := newSetOptions(ioStreaam)
	cmd := &cobra.Command{
		Use:     "set",
		Aliases: []string{},
		Short:   "set local pod ip eBPF map",
		Long:    "set local pod ip eBPF map",
		Example: fmt.Sprintf("    %s localpodips set --pod-ip 10.244.5.100 --ns-index 3 --ns-mac foo --host-index 3 --host-mac bar", name),
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(cmd, args))
			cmdutil.CheckErr(o.Validate(args))
			cmdutil.CheckErr(o.Run())
		},
	}
	cmd.Flags().StringVarP(&o.podIP, "pod-ip", "", o.podIP, "The pod-ip define the eBPF map key")
	cmd.Flags().IntVarP(&o.nsIndex, "ns-index", "", o.nsIndex, "The ns-index define the net ns index")
	cmd.Flags().StringVarP(&o.nsMac, "ns-mac", "", o.nsMac, "The ns-mac define the ns veth pair mac")
	cmd.Flags().IntVarP(&o.hostIndex, "host-index", "", o.hostIndex, "The host-index define the host index")
	cmd.Flags().StringVarP(&o.hostMac, "host-mac", "", o.hostMac, "The host-mac define the host veth pair mac")

	return cmd
}

func (o *setOptions) Complete(cmd *cobra.Command, args []string) error {
	return nil
}

func (o *setOptions) Validate(args []string) error {
	if len(o.podIP) == 0 {
		return fmt.Errorf("pod-ip is required")
	}
	if o.localIpsMap == nil {
		return fmt.Errorf("failed to load eBPF map")
	}
	return nil
}

func (o *setOptions) Run() error {
	if err := o.localIpsMap.Put(
		bpfmap.LocalIpsMapKey{IP: util.InetIpToUInt32(o.podIP)},
		bpfmap.LocalIpsMapInfo{
			IfIndex:    uint32(o.nsIndex),
			LxcIfIndex: uint32(o.hostIndex),
			MAC:        util.Stuff8Byte([]byte(o.nsMac)),
			NodeMAC:    util.Stuff8Byte([]byte(o.hostMac)),
		}); err != nil {
		return fmt.Errorf("failed to set local pod ip %s to map: %w", o.podIP, err)
	}

	fmt.Fprintf(o.Out, "set map successfully %s\n", o.podIP)
	return nil
}
