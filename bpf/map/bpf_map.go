package bpf_map

import (
	"fmt"

	"github.com/cilium/ebpf"
)

const (
	LocalDev      = "/sys/fs/bpf/local_dev"
	LocalPodIps   = "/sys/fs/bpf/local_pod_ips"
	ClusterPodIps = "/sys/fs/bpf/cluster_pod_ips"
)

var (
	localPodIpsMap   *ebpf.Map
	clusterPodIpsMap *ebpf.Map
	localDevMap      *ebpf.Map
)

func InitLoadPinnedMap() error {
	var err error
	localPodIpsMap, err = ebpf.LoadPinnedMap(LocalPodIps, &ebpf.LoadPinOptions{})
	if err != nil {
		return fmt.Errorf("load map error: %w", err)
	}
	clusterPodIpsMap, err = ebpf.LoadPinnedMap(ClusterPodIps, &ebpf.LoadPinOptions{})
	if err != nil {
		return fmt.Errorf("load map error: %w", err)
	}
	localDevMap, err = ebpf.LoadPinnedMap(LocalDev, &ebpf.LoadPinOptions{})
	if err != nil {
		return fmt.Errorf("load map error: %w", err)
	}
	return nil
}

func GetLocalPodIpsMap() *ebpf.Map {
	if localPodIpsMap == nil {
		_ = InitLoadPinnedMap()
	}
	return localPodIpsMap
}

func GetClusterPodIpsMap() *ebpf.Map {
	if clusterPodIpsMap == nil {
		_ = InitLoadPinnedMap()
	}
	return clusterPodIpsMap
}

func GetLocalDevMap() *ebpf.Map {
	if localDevMap == nil {
		_ = InitLoadPinnedMap()
	}
	return localDevMap
}
