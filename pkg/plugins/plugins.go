package plugins

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
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
	"google.golang.org/grpc/credentials/insecure"

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
	Gateway string `json:"gateway"`
	MTU     int    `json:"mtu"`
}

func loadConfig(bytes []byte) (*PluginConf, error) {
	conf := PluginConf{}
	if err := json.Unmarshal(bytes, &conf); err != nil {
		return nil, err
	}
	return &conf, nil
}

func newAgentClient() (ipamapiv1.IpServiceClient, *grpc.ClientConn, error) {
	conn, err := grpc.Dial(":50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	return ipamapiv1.NewIpServiceClient(conn), conn, nil
}

func isHealthy(health ipamapiv1.HealthyType) bool {
	return health == ipamapiv1.HealthyType_Healthy
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

func setIPForNsPair(nsPair *netlink.Veth, ip string) error {
	ip32 := fmt.Sprintf("%s/%s", ip, "32")
	ipAddr, ipNet, err := net.ParseCIDR(ip32)
	if err != nil {
		return err
	}
	ipNet.IP = ipAddr
	link, err := netlink.LinkByName(nsPair.Name)
	if err != nil {
		return err
	}
	return netlink.AddrAdd(link, &netlink.Addr{IPNet: ipNet})
}

func setIPForHostPair(gwPair *netlink.Veth, ip string) error {
	existIp, _ := nettools.DeviceExistIp(gwPair)
	if len(existIp) != 0 {
		return nil
	}
	ip32 := fmt.Sprintf("%s/%s", ip, "32")
	ipAddr, ipNet, err := net.ParseCIDR(ip32)
	if err != nil {
		return err
	}
	ipNet.IP = ipAddr
	link, err := netlink.LinkByName(gwPair.Name)
	if err != nil {
		return err
	}
	return netlink.AddrAdd(link, &netlink.Addr{IPNet: ipNet})
}

func createNsVethPair(ifname string, mtu int, vethName string) (*netlink.Veth, *netlink.Veth, error) {
	return nettools.CreateVethPair(ifname, mtu, vethName)
}

func setHostPairIntoHost(veth *netlink.Veth, netNs ns.NetNS) error {
	return netlink.LinkSetNsFd(veth, int(netNs.Fd()))
}

func setFibTableIntoNs(veth *netlink.Veth, gw string) error {
	gwIp, gwNet, err := net.ParseCIDR(fmt.Sprintf("%s/32", gw))
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
		Dst:       gwNet,
		Gw:        defIp,
	}); err != nil {
		return err
	}

	return netlink.RouteAdd(&netlink.Route{
		LinkIndex: veth.Attrs().Index,
		Scope:     netlink.SCOPE_UNIVERSE,
		Dst:       defNet,
		Gw:        gwIp,
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

	nsVethPodIp := util.InetIpToUInt32(podIP)
	hostVethIndex := uint32(hostVeth.Attrs().Index)
	hostVethMac := stuff8Byte(([]byte)(hostVeth.Attrs().HardwareAddr))
	nsVethIndex := uint32(nsVeth.Attrs().Index)
	nsVethMac := stuff8Byte(([]byte)(nsVeth.Attrs().HardwareAddr))

	bpfMap := bpfmap.GetLocalPodIpsMap()

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

	pluginConfig, err := loadConfig(args.StdinData)
	if err != nil {
		logger.WithError(err).Error("failed to load plugin config")
		return err
	}

	k8sArgs := K8sArgs{}
	if err := types.LoadArgs(args.Args, &k8sArgs); err != nil {
		err := fmt.Errorf("failed to load CNI ENV args: %w", err)
		logger.WithError(err).Error("get k8s arg error")
		return err
	}

	logger.WithFields(logrus.Fields{
		"ContainerID":  args.ContainerID,
		"NetNs":        args.Netns,
		"IfName":       args.IfName,
		"Args":         args.Args,
		"Path":         args.Path,
		"PodNamespace": string(k8sArgs.K8S_POD_NAMESPACE),
		"PodName":      string(k8sArgs.K8S_POD_NAME),
		"PodUid":       string(k8sArgs.K8S_POD_UID),
	}).Info("ADD")

	agentClient, conn, err := newAgentClient()
	if err != nil {
		logger.WithError(err).Error("failed to new agent client")
		return err
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	hresp, err := agentClient.Health(ctx, &ipamapiv1.HealthRequest{})
	if err != nil {
		return err
	}
	if !isHealthy(hresp.Health) {
		return fmt.Errorf("ipam svervice is unhealthy")
	}

	resp, err := agentClient.Allocate(ctx, &ipamapiv1.AllocateRequest{
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

	gwIP := pluginConfig.Gateway
	if len(gwIP) == 0 {
		return fmt.Errorf("failed to get gatewa ip, please setting for node")
	}

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
	if err := setIPForHostPair(gwPair, gwIP); err != nil {
		logger.WithError(err).Error("failed to set ip for host pair")
		return err
	}

	netNs, err := ns.GetNS(args.Netns)
	if err != nil {
		logger.WithError(err).Error("failed to get netns")
		return err
	}

	var nsPair, hostPair *netlink.Veth
	err = netNs.Do(func(hostNs ns.NetNS) error {
		// create veth pair for netns
		vethName := util.GenerateVethName(
			"fast",
			fmt.Sprintf("%s/%s", string(k8sArgs.K8S_POD_NAMESPACE), string(k8sArgs.K8S_POD_NAME)))
		nsPair, hostPair, err = createNsVethPair(args.IfName, 1450, vethName)
		if err != nil {
			logger.WithError(err).Error("failed to create veth pair")
			return err
		}

		// host pair connect to host
		if err := setHostPairIntoHost(hostPair, hostNs); err != nil {
			logger.WithError(err).Error("failed to set host pair to host")
			return err
		}

		// set ip for ns pair
		if err := setIPForNsPair(nsPair, resp.Ip); err != nil {
			logger.WithError(err).Error("failed to set ip for ns pair")
			return err
		}

		// set up ns pair
		if err := setUpVethPair(nsPair); err != nil {
			logger.WithError(err).Error("failed to up ns pair")
			return err
		}

		// add arp table for ns pair
		if err := setFibTableIntoNs(nsPair, gwIP); err != nil {
			logger.WithError(err).Error("failed to set arp table into ns")
			return err
		}
		if err := setArp(gwIP, hostNs, hostPair, args.IfName); err != nil {
			logger.WithError(err).Error("failed to set arp")
			return err
		}

		// set up host pair
		if err := setUpHostPair(hostNs, hostPair); err != nil {
			logger.WithError(err).Error("failed to up host pair")
			return err
		}

		if err := setVethPairInfoToLocalIPsMap(hostNs, resp.Ip, hostPair, nsPair); err != nil {
			logger.WithError(err).Error("failed to save pod information for local ips map")
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	// attach bpf for veth ingress
	if err := attachTcBPFIntoVeth(hostPair); err != nil {
		logger.WithError(err).Error("failed to attach eBPF program to tc ingress")
		return err
	}

	// crate vxlan
	vxlan, err := nettools.CreateVxlanAndUp("fast-vxlan")
	if err != nil {
		logger.WithError(err).Error("failed to create vxlan and up")
		return err
	}

	// save vxlan information to local map
	if err := setVxlanInfoToLocalDevMap(vxlan); err != nil {
		logger.WithError(err).Error("failed to save vxlan information to local map")
		return err
	}

	if err := attachTcBPFIntoVxlan(vxlan); err != nil {
		logger.WithError(err).Error("failed to attach eBPF program to vxlan")
		return err
	}

	_gw, _, _ := net.ParseCIDR(gwIP)
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
	k8sArgs := K8sArgs{}
	if err := types.LoadArgs(args.Args, &k8sArgs); err != nil {
		err := fmt.Errorf("failed to load CNI ENV args: %w", err)
		logger.WithError(err).Error("get k8s arg error")
		return err
	}

	logger.WithFields(logrus.Fields{
		"ContainerID":  args.ContainerID,
		"NetNs":        args.Netns,
		"IfName":       args.IfName,
		"Args":         args.Args,
		"Path":         args.Path,
		"PodNamespace": string(k8sArgs.K8S_POD_NAMESPACE),
		"PodName":      string(k8sArgs.K8S_POD_NAME),
		"PodUid":       string(k8sArgs.K8S_POD_UID),
	}).Info("DEL")

	agentClient, conn, err := newAgentClient()
	if err != nil {
		logger.WithError(err).Error("failed to new agent client")
		return err
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	hresp, err := agentClient.Health(ctx, &ipamapiv1.HealthRequest{})
	if err != nil {
		return err
	}

	if !isHealthy(hresp.Health) {
		return fmt.Errorf("ipam svervice is unhealthy")
	}

	_, err = agentClient.Release(ctx, &ipamapiv1.AllocateRequest{
		Command:   "DEL",
		Id:        args.ContainerID,
		IfName:    args.IfName,
		Namespace: string(k8sArgs.K8S_POD_NAMESPACE),
		Name:      string(k8sArgs.K8S_POD_NAME),
	})
	if err != nil {
		logger.WithError(err).Error("failed to release ip")
	}

	return nil
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
	skel.PluginMain(cmdAdd, cmdCheck, cmdDel, version.All, bv.BuildString("fast"))
}
