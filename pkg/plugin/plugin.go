package plugin

import (
	"context"
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
		"Add", "ContainerID: ", args.ContainerID,
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
	if hresp.Msg != "ok" {
		return err
	}

	_, err = client.Allocate(ctx, &ipamapiv1.IPAMRequest{
		Command: "Add",
		Id:      args.ContainerID,
		IfName:  args.IfName,
	})
	if err != nil {
		return err
	}

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
	if hresp.Msg != "ok" {
		return err
	}

	_, err = client.Release(ctx, &ipamapiv1.IPAMRequest{
		Command: "Add",
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
