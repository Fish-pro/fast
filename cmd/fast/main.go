package main

import (
	"os"

	bpfmap "github.com/fast-io/fast/bpf/map"
	"github.com/fast-io/fast/pkg/plugins"
)

func main() {
	if err := bpfmap.InitLoadPinnedMap(); err != nil {
		os.Exit(0)
	}
	plugins.Main()
}
