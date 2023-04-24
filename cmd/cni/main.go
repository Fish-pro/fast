package main

import (
	"os"

	bpfmap "github.com/fast-io/fast/bpf/map"
	"github.com/fast-io/fast/pkg/plugins"
	"github.com/fast-io/fast/pkg/util"
)

func main() {
	if err := bpfmap.InitLoadPinnedMap(); err != nil {
		util.WriteLog("main exit", "error: ", err.Error())
		os.Exit(0)
	}
	plugins.Main()
}
