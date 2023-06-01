package main

import (
	"k8s.io/component-base/cli"
	"k8s.io/kubectl/pkg/cmd/util"

	"github.com/fast-io/fast/pkg/fastctl"
)

func main() {
	cmd := fastctl.NewFastCtlCommand("fastctl")
	if err := cli.RunNoErrOutput(cmd); err != nil {
		util.CheckErr(err)
	}
}
