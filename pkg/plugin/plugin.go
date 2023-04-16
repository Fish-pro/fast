package plugin

import (
	"runtime"

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
