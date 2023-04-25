package plugins

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"runtime"
	"time"

	"github.com/cilium/ebpf"
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	current "github.com/containernetworking/cni/pkg/types/100"
	"github.com/containernetworking/cni/pkg/version"
	"github.com/containernetworking/plugins/pkg/ns"
	bv "github.com/containernetworking/plugins/pkg/utils/buildversion"
	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"google.golang.org/grpc"

	ipamapiv1 "github.com/fast-io/fast/api/proto/v1"
	bpfmap "github.com/fast-io/fast/bpf/map"
	"github.com/fast-io/fast/bpf/tc"
	"github.com/fast-io/fast/pkg/nettools"
	"github.com/fast-io/fast/pkg/util"
)

const (
	VethHostName = "veth_host"
	VethNetName  = "veth_net"
)

func init() {
	runtime.LockOSThread()
}

type PluginConf struct {
	types.NetConf
	RuntimeConfig *struct {
		TestConfig map[string]interface{} `json:"testConfig"`
	} `json:"runtimeConfig"`

	Bridge string `json:"bridge"`
	Subnet string `json:"subnet"`
	MTU    int    `json:"mtu"`
}

func loadConfig(bytes []byte) (*PluginConf, error) {
	conf := PluginConf{}
	if err := json.Unmarshal(bytes, &conf); err != nil {
		return nil, err
	}
	return &conf, nil
}

func createHostVethPair() (*netlink.Veth, *netlink.Veth, error) {
	hostVeth, _ := netlink.LinkByName(VethHostName)
	netVeth, _ := netlink.LinkByName(VethNetName)
	if hostVeth != nil && netVeth != nil {
		return hostVeth.(*netlink.Veth), netVeth.(*netlink.Veth), nil
	}
	return nettools.CreateVethPair(VethHostName, 1500, VethNetName)
}

func setUpVethPair(veth ...*netlink.Veth) error {
	for _, v := range veth {
		if err := netlink.LinkSetUp(v); err != nil {
			return err
		}
	}
	return nil
}

func setIPForPair(name string, ip string) error {
	ip32 := fmt.Sprintf("%s/%s", ip, "32")
	ipAddr, ipNet, err := net.ParseCIDR(ip32)
	ipNet.IP = ipAddr
	link, err := netlink.LinkByName(name)
	if err != nil {
		return err
	}
	return netlink.AddrAdd(link, &netlink.Addr{IPNet: ipNet})
}

func createNsVethPair(ifname string, mtu int) (*netlink.Veth, *netlink.Veth, error) {
	return nettools.CreateVethPair(ifname, mtu)
}

func setHostPairIntoHost(veth *netlink.Veth, netNs ns.NetNS) error {
	return netlink.LinkSetNsFd(veth, int(netNs.Fd()))
}

func setFibTableIntoNs(veth *netlink.Veth, gw string) error {
	gwIp, gwNet, err := net.ParseCIDR(gw)
	if err != nil {
		return err
	}
	defIp, defNet, err := net.ParseCIDR("0.0.0.0/0")
	if err != nil {
		return err
	}

	if err := netlink.RouteAdd(&netlink.Route{
		LinkIndex: veth.Attrs().Index,
		Scope:     netlink.SCOPE_LINK,
		Dst:       defNet,
		Gw:        gwIp,
	}); err != nil {
		return err
	}

	return netlink.RouteAdd(&netlink.Route{
		LinkIndex: veth.Attrs().Index,
		Scope:     netlink.SCOPE_UNIVERSE,
		Dst:       gwNet,
		Gw:        defIp,
	})
}

func setArp(gwIP string, hostNs ns.NetNS, veth *netlink.Veth, dev string) error {
	return hostNs.Do(func(netNs ns.NetNS) error {
		v, err := netlink.LinkByName(veth.Attrs().Name)
		if err != nil {
			return err
		}
		veth := v.(*netlink.Veth)
		mac := veth.LinkAttrs.HardwareAddr
		_mac := mac.String()
		return netNs.Do(func(hostNs ns.NetNS) error {
			return nettools.CreateArpEntry(gwIP, _mac, dev)
		})
	})
}

func setUpHostPair(hostNs ns.NetNS, veth *netlink.Veth) error {
	return hostNs.Do(func(netNS ns.NetNS) error {
		v, err := netlink.LinkByName(veth.Attrs().Name)
		if err != nil {
			return err
		}
		veth := v.(*netlink.Veth)
		return setUpVethPair(veth)
	})
}

func setVethPairInfoToLocalIPsMap(hostNs ns.NetNS, podIP string, hostVeth, nsVeth *netlink.Veth) error {
	err := hostNs.Do(func(nn ns.NetNS) error {
		v, err := netlink.LinkByName(hostVeth.Attrs().Name)
		if err != nil {
			return err
		}
		hostVeth = v.(*netlink.Veth)
		return nil
	})
	if err != nil {
		return err
	}
	netip, _, err := net.ParseCIDR(podIP)
	if err != nil {
		return err
	}
	podIP = netip.String()

	nsVethPodIp := util.InetIpToUInt32(podIP)
	hostVethIndex := uint32(hostVeth.Attrs().Index)
	hostVethMac := stuff8Byte(([]byte)(hostVeth.Attrs().HardwareAddr))
	nsVethIndex := uint32(nsVeth.Attrs().Index)
	nsVethMac := stuff8Byte(([]byte)(nsVeth.Attrs().HardwareAddr))

	bpfMap := bpfmap.GetLocalDevMap()

	return bpfMap.Update(
		bpfmap.LocalIpsMapKey{IP: nsVethPodIp},
		bpfmap.LocalIpsMapInfo{
			IfIndex:    nsVethIndex,
			LxcIfIndex: hostVethIndex,
			MAC:        nsVethMac,
			NodeMAC:    hostVethMac,
		},
		ebpf.UpdateAny,
	)
}

func stuff8Byte(b []byte) [8]byte {
	var res [8]byte
	if len(b) > 8 {
		b = b[0:9]
	}

	for index, _byte := range b {
		res[index] = _byte
	}
	return res
}

func attachTcBPFIntoVeth(veth *netlink.Veth) error {
	name := veth.Attrs().Name
	vethIngressBPFPath := tc.GetVethIngressPath()
	return tc.TryAttachBPF(name, tc.IngressType, vethIngressBPFPath)
}

func setVxlanInfoToLocalDevMap(vxlan *netlink.Vxlan) error {
	bpfMap := bpfmap.GetLocalDevMap()
	return bpfMap.Update(
		bpfmap.LocalDevMapKey{
			Type: bpfmap.VxlanDevType,
		},
		bpfmap.LocalDevMapValue{
			IfIndex: uint32(vxlan.Attrs().Index),
		},
		ebpf.UpdateAny,
	)
}

func attachTcBPFIntoVxlan(vxlan *netlink.Vxlan) error {
	name := vxlan.Attrs().Name
	vxlanIngressBPFPath := tc.GetVxlanIngressPath()
	err := tc.TryAttachBPF(name, tc.IngressType, vxlanIngressBPFPath)
	if err != nil {
		return err
	}
	vxlanEgressBPFPath := tc.GetVxlanEgressPath()
	return tc.TryAttachBPF(name, tc.EgressType, vxlanEgressBPFPath)
}

/*
 * tc qdisc add dev ${pod veth name} clsact
 * tc qdisc add dev fast_vxlan clsact
 * clang -g  -O2 -emit-llvm -c vxlan_egress.c -o - | llc -march=bpf -filetype=obj -o vxlan_egress.o
 * clang -g  -O2 -emit-llvm -c vxlan_ingress.c -o - | llc -march=bpf -filetype=obj -o vxlan_ingress.o
 * clang -g  -O2 -emit-llvm -c veth_ingress.c -o - | llc -march=bpf -filetype=obj -o veth_ingress.o
 * tc filter add dev fast_vxlan egress bpf direct-action obj vxlan_egress.o
 * tc filter add dev fast_vxlan ingress bpf direct-action obj vxlan_ingress.o
 * tc filter add dev ${pod veth name} ingress bpf direct-action obj veth_ingress.o
 */
func cmdAdd(args *skel.CmdArgs) error {
	logger := util.NewLogger()
	logger.WithFields(logrus.Fields{
		"ContainerID": args.ContainerID,
		"NetNs":       args.Netns,
		"IfName":      args.IfName,
		"Args":        args.Args,
		"Path":        args.Path,
		"StdinData":   string(args.StdinData),
	}).Info("ADD")

	k8sArgs := K8sArgs{}
	if err := types.LoadArgs(args.Args, &k8sArgs); err != nil {
		err := fmt.Errorf("failed to load CNI ENV args: %v", err)
		logger.WithError(err).Error("get k8s arg error")
		return err
	}

	logger.WithFields(logrus.Fields{
		"PodNamespace": string(k8sArgs.K8S_POD_NAMESPACE),
		"PodName":      string(k8sArgs.K8S_POD_NAME),
		"PodUid":       string(k8sArgs.K8S_POD_UID),
	}).Info("get k8s args")

	pluginConfig, err := loadConfig(args.StdinData)
	if err != nil {
		logger.WithError(err).Error("failed to load plugin config")
		return err
	}

	conn, err := grpc.Dial(":8999")
	if err != nil {
		logger.WithError(err).Error("failed to connect server")
		return err
	}
	defer conn.Close()

	client := ipamapiv1.NewIpServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	hresp, err := client.Health(ctx, &ipamapiv1.HealthRequest{})
	if err != nil {
		return err
	}
	if !util.IsHealthy(hresp) {
		return fmt.Errorf("ipam svervice is unhealthy")
	}

	resp, err := client.Allocate(ctx, &ipamapiv1.AllocateRequest{
		Command:   "ADD",
		Id:        args.ContainerID,
		IfName:    args.IfName,
		Namespace: string(k8sArgs.K8S_POD_NAMESPACE),
		Name:      string(k8sArgs.K8S_POD_NAME),
		Uid:       string(k8sArgs.K8S_POD_UID),
	})
	if err != nil {
		return err
	}
	logger.WithFields(logrus.Fields{
		"namespace": string(k8sArgs.K8S_POD_NAMESPACE),
		"name":      string(k8sArgs.K8S_POD_NAME),
		"ip":        resp.Ip,
	}).Info("allocate ip successfully")

	hostName, err := os.Hostname()
	if err != nil {
		return err
	}

	gwResp, err := client.GetGateway(ctx, &ipamapiv1.GatewayRequest{
		Node: hostName,
	})

	// create or get veth_host and veth_net
	gwPair, netPair, err := createHostVethPair()
	if err != nil {
		logger.WithError(err).Error("failed to create host veth pair")
		return err
	}

	// set up host veth pair
	if err := setUpVethPair(gwPair, netPair); err != nil {
		logger.WithError(err).Error("failed to set up host veth pair")
		return err
	}

	// set ip for host pair
	if err := setIPForPair(gwPair.Name, gwResp.Gateway); err != nil {
		logger.WithError(err).Error("failed to set ip for host pair")
		return err
	}

	netNs, err := ns.GetNS(args.Netns)
	if err != nil {
		return err
	}

	var nsPair, hostPair *netlink.Veth
	err = netNs.Do(func(hostNs ns.NetNS) error {
		// create veth pair for netns
		nsPair, hostPair, err = createNsVethPair(args.IfName, 1450)
		if err != nil {
			return err
		}

		// host pair connect to host
		if err := setHostPairIntoHost(hostPair, hostNs); err != nil {
			return err
		}

		// set ip for ns pair
		if err := setIPForPair(nsPair.Name, resp.Ip); err != nil {
			return err
		}

		// set up ns pair
		if err := setUpVethPair(nsPair); err != nil {
			return err
		}

		// add arp table for ns pair
		if err := setFibTableIntoNs(nsPair, gwResp.Gateway); err != nil {
			return err
		}
		if err := setArp(gwResp.Gateway, hostNs, hostPair, args.IfName); err != nil {
			return err
		}

		// set up host pair
		if err := setUpHostPair(hostNs, hostPair); err != nil {
			return err
		}

		return setVethPairInfoToLocalIPsMap(hostNs, resp.Ip, hostPair, nsPair)
	})
	if err != nil {
		return err
	}

	// attach bpf for veth ingress
	if err := attachTcBPFIntoVeth(hostPair); err != nil {
		return err
	}

	// crate vxlan
	vxlan, err := nettools.CreateVxlanAndUp("fast-vxlan")
	if err != nil {
		return err
	}

	// save vxlan information to local map
	if err := setVxlanInfoToLocalDevMap(vxlan); err != nil {
		return err
	}

	if err := attachTcBPFIntoVxlan(vxlan); err != nil {
		return err
	}

	_gw, _, _ := net.ParseCIDR(gwResp.Gateway)
	_, _podIP, _ := net.ParseCIDR(resp.Ip)
	result := &current.Result{
		CNIVersion: pluginConfig.CNIVersion,
		IPs: []*current.IPConfig{
			{
				Address: *_podIP,
				Gateway: _gw,
			},
		},
	}
	return types.PrintResult(result, pluginConfig.CNIVersion)
}

func cmdDel(args *skel.CmdArgs) error {
	logger := util.NewLogger()
	logger.WithFields(logrus.Fields{
		"ContainerID": args.ContainerID,
		"NetNs":       args.Netns,
		"IfName":      args.IfName,
		"Args":        args.Args,
		"Path":        args.Path,
		"StdinData":   string(args.StdinData),
	}).Info("DEL")
	conn, err := grpc.Dial(":8999")
	if err != nil {
		logger.WithError(err).Error("failed to connect grpc server")
		return err
	}
	defer conn.Close()

	agentClient := ipamapiv1.NewIpServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	hresp, err := agentClient.Health(ctx, &ipamapiv1.HealthRequest{})
	if err != nil {
		return err
	}

	if !util.IsHealthy(hresp) {
		return fmt.Errorf("ipam svervice is unhealthy")
	}

	_, err = agentClient.Release(ctx, &ipamapiv1.AllocateRequest{
		Command: "DEL",
		Id:      args.ContainerID,
		IfName:  args.IfName,
	})

	return err
}

func cmdCheck(args *skel.CmdArgs) error {
	logger := util.NewLogger()
	logger.WithFields(logrus.Fields{
		"ContainerID": args.ContainerID,
		"NetNs":       args.Netns,
		"IfName":      args.IfName,
		"Args":        args.Args,
		"Path":        args.Path,
		"StdinData":   string(args.StdinData),
	}).Info("CHECK")
	return nil
}

func Main() {
	skel.PluginMain(cmdAdd, cmdCheck, cmdDel, version.All, bv.BuildString("fastcni"))
}
