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
	"net"
	"os"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
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

	ipamapiv1 "github.com/fast-io/fast/api/proto/v1"
	ipamservicev1 "github.com/fast-io/fast/api/service/v1"
	bpfmap "github.com/fast-io/fast/bpf/map"
	"github.com/fast-io/fast/cmd/agent/app/config"
	"github.com/fast-io/fast/cmd/agent/app/options"
	clientbuilder "github.com/fast-io/fast/pkg/builder"
	clusterpodctrl "github.com/fast-io/fast/pkg/controllers/clusterpod"
	ipsinformers "github.com/fast-io/fast/pkg/generated/informers/externalversions"
)

func init() {
	utilruntime.Must(logsapi.AddFeatureGates(utilfeature.DefaultMutableFeatureGate))
	utilruntime.Must(features.AddFeatureGates(utilfeature.DefaultMutableFeatureGate))
}

// NewAgentCommand returns the agent root command
func NewAgentCommand() *cobra.Command {
	o := options.NewAgentOptions()

	cmd := &cobra.Command{
		Use:  "fast-agent",
		Long: `The fast-agent is an agent component`,
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

	clientBuilder := clientbuilder.NewSimpleIpsControllerClientBuilder(c.Kubeconfig)
	ipsManager := clientBuilder.IpsClientOrDie("ips-manager")

	// 1.create map and attach eBPF programs
	if err := bpfmap.InitLoadPinnedMap(); err != nil {
		return err
	}

	// new normal informer factory
	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(c.Client, time.Second*30)
	// new ips informer factory
	ipsInformerFactory := ipsinformers.NewSharedInformerFactory(ipsManager, time.Second*30)

	// 2.Obtain the cluster pod IP and store the information to the cluster eBPF map
	controller, err := clusterpodctrl.NewController(
		ctx,
		c.Client,
		kubeInformerFactory.Core().V1().Pods(),
	)
	if err != nil {
		return err
	}
	go controller.Run(ctx)

	// 3.start grpc server
	var server *grpc.Server
	var opts []grpc.ServerOption
	server = grpc.NewServer(opts...)
	listen, err := net.Listen("tcp", ":"+"8999")
	if err != nil {
		logger.Error(err, "gRPC listen error")
		return err
	}
	ipamSvc := ipamservicev1.NewIPAMService(
		ctx,
		logger,
		clientBuilder.IpsClientOrDie("fast-agent"),
		kubeInformerFactory.Core().V1().Pods(),
		ipsInformerFactory.Sample().V1alpha1().Ipses(),
		ipsInformerFactory.Sample().V1alpha1().IpEndpoints(),
	)
	ipamapiv1.RegisterIpServiceServer(server, ipamSvc)

	go func() {
		logger.Info("starting gRPC server...")
		err = server.Serve(listen)
		if err != nil {
			logger.Error(err, "start gRPC server error")
		}
	}()

	kubeInformerFactory.Start(stopCh)
	ipsInformerFactory.Start(stopCh)

	<-stopCh
	return nil
}
