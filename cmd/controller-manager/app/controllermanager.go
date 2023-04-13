/*
Copyright 2023 The ips Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package app

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	kubeinformers "k8s.io/client-go/informers"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/cli/globalflag"
	logsapi "k8s.io/component-base/logs/api/v1"
	"k8s.io/component-base/metrics/features"
	"k8s.io/component-base/term"
	"k8s.io/component-base/version"
	"k8s.io/component-base/version/verflag"
	"k8s.io/klog/v2"

	"github.com/fast-io/fast/cmd/controller-manager/app/config"
	"github.com/fast-io/fast/cmd/controller-manager/app/options"
	ipsctrl "github.com/fast-io/fast/pkg/controllers/ips"
	ipsinformers "github.com/fast-io/fast/pkg/generated/informers/externalversions"
)

func init() {
	utilruntime.Must(logsapi.AddFeatureGates(utilfeature.DefaultMutableFeatureGate))
	utilruntime.Must(features.AddFeatureGates(utilfeature.DefaultMutableFeatureGate))
}

// NewControllerManagerCommand returns the agent root command
func NewControllerManagerCommand() *cobra.Command {
	o := options.NewAgentOptions()

	cmd := &cobra.Command{
		Use:  "fast-controller-manger",
		Long: `The fast-controller-manger is an controller manager`,
		RunE: func(cmd *cobra.Command, args []string) error {
			verflag.PrintAndExitIfRequested()
			// Activate logging as soon as possible, after that
			// show flags with the final logging configuration.
			if err := logsapi.ValidateAndApply(o.Logs, utilfeature.DefaultFeatureGate); err != nil {
				return err
			}
			cliflag.PrintFlags(cmd.Flags())

			c, err := o.Config()
			if err != nil {
				return err
			}
			// add feature enablement metrics
			utilfeature.DefaultMutableFeatureGate.AddMetrics()
			return Run(context.Background(), c.Complete())
		},
		Args: func(cmd *cobra.Command, args []string) error {
			for _, arg := range args {
				if len(arg) > 0 {
					return fmt.Errorf("%q does not take any arguments, got %q", cmd.CommandPath(), args)
				}
			}
			return nil
		},
	}

	fs := cmd.Flags()
	namedFlagSets := o.Flags()
	verflag.AddFlags(namedFlagSets.FlagSet("global"))
	globalflag.AddGlobalFlags(namedFlagSets.FlagSet("global"), cmd.Name())
	for _, f := range namedFlagSets.FlagSets {
		fs.AddFlagSet(f)
	}

	cols, _, _ := term.TerminalSize(cmd.OutOrStdout())
	cliflag.SetUsageAndHelpFunc(cmd, namedFlagSets, cols)

	return cmd
}

// Run runs the agent controller and attach ebpf program. This should never exit.
func Run(ctx context.Context, c *config.CompletedConfig) error {
	logger := klog.FromContext(ctx)
	stopCh := ctx.Done()

	// To help debugging, immediately log version
	logger.Info("Starting", "version", version.Get())

	logger.Info("Golang settings", "GOGC", os.Getenv("GOGC"), "GOMAXPROCS", os.Getenv("GOMAXPROCS"), "GOTRACEBACK", os.Getenv("GOTRACEBACK"))

	// Start events processing pipeline.
	c.EventBroadcaster.StartStructuredLogging(0)
	c.EventBroadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: c.Client.CoreV1().Events("")})
	defer c.EventBroadcaster.Shutdown()

	// new normal informer factory
	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(c.Client, time.Second*30)
	// new ips informer factory
	ipsInformerFactory := ipsinformers.NewSharedInformerFactory(c.IClient, time.Second*30)

	// ips controller
	ipsController, err := ipsctrl.NewController(
		ctx,
		c.Client,
		c.IClient,
		ipsInformerFactory.Sample().V1alpha1().Ipses())
	if err != nil {
		logger.Error(err, "Failed to new ips controller")
		return err
	}
	go ipsController.Run(ctx)

	// gc controller

	kubeInformerFactory.Start(stopCh)
	ipsInformerFactory.Start(stopCh)

	<-stopCh
	return nil
}
