package bpf_map

import (
	"fmt"
	"unsafe"

	"github.com/cilium/ebpf"
)

const (
	LocalDev      = "/sys/fs/bpf/tc/globals/local_dev"
	LocalPodIps   = "/sys/fs/bpf/tc/globals/local_pod_ips"
	ClusterPodIps = "/sys/fs/bpf/tc/globals/cluster_pod_ips"
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

func PrintMapSize() {
	fmt.Println(uint32(unsafe.Sizeof(LocalDevMapKey{})))
	fmt.Println(uint32(unsafe.Sizeof(LocalDevMapValue{})))
	fmt.Println(uint32(unsafe.Sizeof(LocalIpsMapKey{})))
	fmt.Println(uint32(unsafe.Sizeof(LocalIpsMapInfo{})))
	fmt.Println(uint32(unsafe.Sizeof(ClusterIpsMapKey{})))
	fmt.Println(uint32(unsafe.Sizeof(ClusterIpsMapInfo{})))
}
