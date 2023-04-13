package main

import (
	"os"

	"k8s.io/component-base/cli"
	_ "k8s.io/component-base/logs/json/register"
	_ "k8s.io/component-base/metrics/prometheus/clientgo"
	_ "k8s.io/component-base/metrics/prometheus/version"

	"github.com/fast-io/fast/cmd/controller-manager/app"
	_ "github.com/fast-io/fast/pkg/apis/ips/install"
)

func main() {
	command := app.NewControllerManagerCommand()
	code := cli.Run(command)
	os.Exit(code)
}
