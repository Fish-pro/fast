package plugins

import (
	"context"
	"fmt"
	ipamapiv1 "github.com/fast-io/fast/api/proto/v1"
	"google.golang.org/grpc"
	"runtime"
	"time"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/version"
	bv "github.com/containernetworking/plugins/pkg/utils/buildversion"

	"github.com/fast-io/fast/pkg/util"
)

type PluginConf struct {
	types.NetConf
	RuntimeConfig *struct {
		TestConfig map[string]interface{} `json:"testConfig"`
	} `json:"runtimeConfig"`

	Bridge string `json:"bridge"`
	Subnet string `json:"subnet"`
	MTU    int    `json:"mtu"`
}

func init() {
	runtime.LockOSThread()
}

func cmdAdd(args *skel.CmdArgs) error {
	util.WriteLog(
		"ADD", "ContainerID: ", args.ContainerID,
		"NetNs: ", args.Netns,
		"IfName: ", args.IfName,
		"Args: ", args.Args,
		"Path: ", args.Path,
		"StdinData: ", string(args.StdinData))

	k8sArgs := K8sArgs{}
	if err := types.LoadArgs(args.Args, &k8sArgs); err != nil {
		err := fmt.Errorf("failed to load CNI ENV args: %v", err)
		util.WriteLog("get k8s arg error", "err: ", err.Error())
		return err
	}

	util.WriteLog(
		"k8s arg: ",
		"PodNamespace", string(k8sArgs.K8S_POD_NAMESPACE),
		"PodName", string(k8sArgs.K8S_POD_NAME),
		"PodUid", string(k8sArgs.K8S_POD_UID),
	)

	conn, err := grpc.Dial(":8999")
	if err != nil {
		util.WriteLog("failed to connect server", "error", err.Error())
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
	util.WriteLog(
		"ADD",
		"namespace: ", string(k8sArgs.K8S_POD_NAMESPACE),
		"name: ", string(k8sArgs.K8S_POD_NAME),
		"allocated ip: ", resp.Ip,
	)

	return nil
}

func cmdDel(args *skel.CmdArgs) error {
	util.WriteLog(
		"Del", "ContainerID: ", args.ContainerID,
		"NetNs: ", args.Netns,
		"IfName: ", args.IfName,
		"Args: ", args.Args,
		"Path: ", args.Path,
		"StdinData: ", string(args.StdinData))
	conn, err := grpc.Dial(":8999")
	if err != nil {
		util.WriteLog("failed to connect server", "error", err.Error())
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

	_, err = client.Release(ctx, &ipamapiv1.AllocateRequest{
		Command: "Del",
		Id:      args.ContainerID,
		IfName:  args.IfName,
	})
	if err != nil {
		return err
	}

	return nil
}

func cmdCheck(args *skel.CmdArgs) error {
	util.WriteLog(
		"Check", "ContainerID: ", args.ContainerID,
		"NetNs: ", args.Netns,
		"IfName: ", args.IfName,
		"Args: ", args.Args,
		"Path: ", args.Path,
		"StdinData: ", string(args.StdinData))
	return nil
}

func Main() {
	skel.PluginMain(cmdAdd, cmdCheck, cmdDel, version.All, bv.BuildString("fastcni"))
}
