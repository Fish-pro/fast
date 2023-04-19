package v1

import (
	"context"

	coreinformers "k8s.io/client-go/informers/core/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"

	apiv1 "github.com/fast-io/fast/api/proto/v1"
	ipsversioned "github.com/fast-io/fast/pkg/generated/clientset/versioned"
	ipsinformers "github.com/fast-io/fast/pkg/generated/informers/externalversions/ips/v1alpha1"
	ipslisters "github.com/fast-io/fast/pkg/generated/listers/ips/v1alpha1"
)

type IPAMService struct {
	client ipsversioned.Interface

	podLister  corelisters.PodLister
	ipsLister  ipslisters.IpsLister
	ipepLister ipslisters.IpEndpointLister

	podSynced  cache.InformerSynced
	ipsSynced  cache.InformerSynced
	ipepSynced cache.InformerSynced

	apiv1.UnimplementedIpServiceServer
}

func NewIPAMService(
	client ipsversioned.Interface,
	podInformer coreinformers.PodInformer,
	ipsInformer ipsinformers.IpsInformer,
	ipepInformer ipsinformers.IpEndpointInformer) apiv1.IpServiceServer {

	return &IPAMService{
		client:     client,
		podLister:  podInformer.Lister(),
		ipsLister:  ipsInformer.Lister(),
		ipepLister: ipepInformer.Lister(),
		podSynced:  podInformer.Informer().HasSynced,
		ipsSynced:  ipsInformer.Informer().HasSynced,
		ipepSynced: ipepInformer.Informer().HasSynced,
	}
}

func (s *IPAMService) Start(ctx context.Context) {
	if !cache.WaitForCacheSync(ctx.Done(), s.podSynced, s.ipsSynced) {
		return
	}
}

func (s *IPAMService) Allocate(context.Context, *apiv1.IPAMRequest) (*apiv1.IPAMResponse, error) {
	return nil, nil
}

func (s *IPAMService) Release(context.Context, *apiv1.IPAMRequest) (*apiv1.IPAMResponse, error) {
	return nil, nil
}
