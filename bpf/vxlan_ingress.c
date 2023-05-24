#include <linux/bpf.h>
#include <linux/pkt_cls.h>
#include <bpf/bpf_helpers.h>
#include <linux/if_ether.h>
#include <linux/ip.h>
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
  bpf_printk("the dst_ip is: %d", dst_ip);
  bpf_printk("the ip->daddr is: %d", ip->daddr);

  struct endpointKey epKey = {};
  epKey.ip = dst_ip;
  struct endpointInfo *ep = bpf_map_lookup_elem(&ding_lxc, &epKey);
  if (!ep) {
    return TC_ACT_OK;
  }
  __u8 src_mac[ETH_ALEN];
	__u8 dst_mac[ETH_ALEN];
  bpf_memcpy(src_mac, ep->nodeMac, ETH_ALEN);
  bpf_memcpy(dst_mac, ep->mac, ETH_ALEN);
  bpf_skb_store_bytes(
    skb,
    offsetof(struct ethhdr, h_dest),
    dst_mac,
    ETH_ALEN,
    0
  );
    bpf_skb_store_bytes(
    skb,
    offsetof(struct ethhdr, h_source),
    src_mac,
    ETH_ALEN,
    0
  );
 
  return bpf_redirect(ep->lxcIfIndex, 0);
}

char _license[] SEC("license") = "GPL";
