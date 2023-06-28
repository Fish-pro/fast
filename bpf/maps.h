#include <linux/bpf.h>
#include <linux/pkt_cls.h>
#include <bpf/bpf_helpers.h>

#include "common.h"

#define LOCAL_DEV_VXLAN 1;
#define LOCAL_DEV_VETH 2;

#define DEFAULT_TUNNEL_ID 13190

struct localIpsMapKey {
  __u32 ip;
};

struct localIpsMapInfo {
  __u32 ifIndex;
  __u32 lxcIfIndex;
  __u8 mac[8];
  __u8 nodeMac[8];
};

// The container IP address of the local node is stored
struct {
  __uint(type, BPF_MAP_TYPE_HASH);
  __uint(max_entries, 255);
  __type(key, struct localIpsMapKey);
  __type(value, struct localIpsMapInfo);
  __uint(pinning, LIBBPF_PIN_BY_NAME);
} local_pod_ip __section_maps_btf;


struct clusterIpsMapKey {
  __u32 ip;
};

struct clusterIpsMapInfo {
  __u32 ip;
};

// The container IP addresses of other nodes are stored
struct {
  __uint(type, BPF_MAP_TYPE_HASH);
  __uint(max_entries, 255);
  __type(key, struct clusterIpsMapKey);
  __type(value, struct clusterIpsMapInfo);
  __uint(pinning, LIBBPF_PIN_BY_NAME);
} cluster_pod_ip __section_maps_btf;

struct localDevMapKey {
  __u32 type;
};

struct localDevMapValue {
  __u32 ifIndex;
};

// Stores the vxlan NIC information
struct {
  __uint(type, BPF_MAP_TYPE_HASH);
  __uint(max_entries, 255);
  __type(key, struct localDevMapKey);
  __type(value, struct localDevMapValue);
  __uint(pinning, LIBBPF_PIN_BY_NAME);
} local_dev __section_maps_btf;

