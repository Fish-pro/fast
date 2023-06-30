#include <linux/bpf.h>
#include <linux/pkt_cls.h>
#include <bpf/bpf_helpers.h>
#include <linux/if_ether.h>
#include <linux/ip.h>
#include <linux/icmp.h>
#include <netinet/in.h>

#include "common.h"
#include "maps.h"

__section("classifier")
int cls_main(struct __sk_buff *skb) {
  void *data = (void *)(long)skb->data;
  void *data_end = (void *)(long)skb->data_end;
  if (data + sizeof(struct ethhdr) + sizeof(struct iphdr) > data_end) {
    return TC_ACT_UNSPEC;
  }

  struct ethhdr  *eth  = data;
  struct iphdr   *ip   = (data + sizeof(struct ethhdr));
  if (eth->h_proto != __constant_htons(ETH_P_IP)) {
		return TC_ACT_UNSPEC;
  }

  __u32 src_ip = htonl(ip->saddr);
  __u32 dst_ip = htonl(ip->daddr);
  __u8 src_mac[ETH_ALEN];
  __u8 dst_mac[ETH_ALEN];
  struct localIpsMapKey epKey = {};
  epKey.ip = dst_ip;
  struct localIpsMapInfo *ep = bpf_map_lookup_elem(&local_pod_ips, &epKey);
  // If the obtained IP address is the IP address of the local node
  if (ep) {
    bpf_memcpy(src_mac, ep->nodeMac, ETH_ALEN);
	bpf_memcpy(dst_mac, ep->mac, ETH_ALEN);
    bpf_skb_store_bytes(skb, offsetof(struct ethhdr, h_source), dst_mac, ETH_ALEN, 0);
	bpf_skb_store_bytes(skb, offsetof(struct ethhdr, h_dest), src_mac, ETH_ALEN, 0);
	// It is directly redirected to the network adapter of the cluster container
    return bpf_redirect_peer(ep->lxcIfIndex, 0);
  }
  struct clusterIpsMapKey podNodeKey = {};
  podNodeKey.ip = dst_ip;
  struct clusterIpsMapInfo *podNode = bpf_map_lookup_elem(&cluster_pod_ips, &podNodeKey);
  // If it is the IP address of another node container
  if (podNode) {
    struct localDevMapKey localKey = {};
    localKey.type = LOCAL_DEV_VXLAN;
    struct localDevMapValue *localValue = bpf_map_lookup_elem(&local_dev, &localKey);
    if (localValue) {
      // Redirect to vxlan
      return bpf_redirect(localValue->ifIndex, 0);
    } 
    return TC_ACT_UNSPEC;
  }
  return TC_ACT_UNSPEC;
}

char _license[] SEC("license") = "GPL";
