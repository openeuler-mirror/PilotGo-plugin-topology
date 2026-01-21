#include <stddef.h>
#include <linux/if_packet.h>
#include <linux/ip.h>
#include <linux/ipv6.h>
#include <linux/icmp.h>
#include <linux/icmpv6.h>
#include <linux/udp.h>
#include <linux/tcp.h>
#include <linux/in.h>

#include "defs.h"
#include "flow.h"
#include "common.h"

MAP(u64_config_values){
	.type = BPF_MAP_TYPE_ARRAY,
	.key_size = sizeof(__u32),
	.value_size = sizeof(__u64),
	.max_entries = 2,
};

MAP(stats_map){
	.type = BPF_MAP_TYPE_HASH,
	.key_size = sizeof(__u32),
	.value_size = sizeof(__u64),
	.max_entries = 1,
};

MAP(flow_table_p1){
	.type = BPF_MAP_TYPE_HASH,
	.key_size = sizeof(__u64),
	.value_size = sizeof(struct flow),
	.max_entries = 500000,
};

MAP(flow_table_p2){
	.type = BPF_MAP_TYPE_HASH,
	.key_size = sizeof(__u64),
	.value_size = sizeof(struct flow),
	.max_entries = 500000,
};

static inline void fill_network(struct __sk_buff *skb, __u16 netproto, int offset,
								struct flow *flow, __u64 tm)
{
	struct network_layer *layer = &flow->network_layer;

	int len = skb->len - sizeof(struct ethhdr);
	int frag = 0;

	__u8 transproto = 0;

	__u64 hash_src = 0;
	__u64 hash_dst = 0;

	__u64 ordered_src = 0;
	__u64 ordered_dst = 0;

#define IP_SRC (__u64) layer->ip_src
#define IP_DST (__u64) layer->ip_dst

	layer->protocol = netproto;
	switch (netproto)
	{
	case ETH_P_IP:
		frag = load_half(skb, offset + offsetof(struct iphdr, frag_off)) & (IP_MF | IP_OFFSET);
		if (frag)
		{
			// TODO report fragment
			return;
		}

		transproto = load_byte(skb, offset + offsetof(struct iphdr, protocol));
		
		fill_ipv4(skb, offset + offsetof(struct iphdr, saddr), layer->ip_src, &hash_src);
		fill_ipv4(skb, offset + offsetof(struct iphdr, daddr), layer->ip_dst, &hash_dst);

		ordered_src = IP_SRC[12] << 24 | IP_SRC[13] << 16 | IP_SRC[14] << 8 | IP_SRC[15];
		ordered_dst = IP_DST[12] << 24 | IP_DST[13] << 16 | IP_DST[14] << 8 | IP_DST[15];
		break;
	case ETH_P_IPV6:
		transproto = load_byte(skb, offset + offsetof(struct ipv6hdr, nexthdr));
		fill_ipv6(skb, offset + offsetof(struct ipv6hdr, saddr), layer->ip_src, &hash_src);
		fill_ipv6(skb, offset + offsetof(struct ipv6hdr, daddr), layer->ip_dst, &hash_dst);

#ifdef FIX_STACK_512LIMIT
		ordered_src = (IP_SRC[0] << 56 | IP_SRC[1] << 48 | IP_SRC[2] << 40 | IP_SRC[3] << 32 |
					   IP_SRC[4] << 24 | IP_SRC[5] << 16 | IP_SRC[6] << 8 | IP_SRC[7]) ^
					  (IP_SRC[8] << 56 | IP_SRC[9] << 48 | IP_SRC[10] << 40 | IP_SRC[11] << 32 |
					   IP_SRC[12] << 24 | IP_SRC[13] << 16 | IP_SRC[14] << 8 | IP_SRC[15]);
		ordered_dst = (IP_DST[0] << 56 | IP_DST[1] << 48 | IP_DST[2] << 40 | IP_DST[3] << 32 |
					   IP_DST[4] << 24 | IP_DST[5] << 16 | Ip_DST[6] << 8 | IP_DST[7]) ^
					  (IP_DST[8] << 56 | IP_DST[9] << 48 | IP_DST[10] << 40 | IP_DST[11] << 32 |
					   IP_DST[12] << 24 | IP_DST[13] << 16 | IP_DST[14] << 8 | IP_DST[15]);
#endif
		break;
	default:
		return;
	}

	__u8 verlen = load_byte(skb, offset);
	offset += (verlen & 0xF) << 2;
	len -= (verlen & 0xF) << 2;

	switch (transproto)
	{
	case IPPROTO_GRE:
		// TODO
		break;
	case IPPROTO_SCTP:
		// TODO
	case IPPROTO_UDP:
	case IPPROTO_TCP:
		fill_transport(skb, transproto, offset, len, flow, ordered_src < ordered_dst, ordered_src == ordered_dst);
		break;
	case IPPROTO_ICMP:
		fill_icmpv4(skb, offset, flow);
		break;
	case IPPROTO_ICMPV6:
		fill_icmpv6(skb, offset, flow);
		break;
	}

	layer->_hash_src = hash_src;
	if (ordered_src < ordered_dst)
	{
		layer->_hash = FNV_BASIS ^ rotl(hash_src, 32) ^ hash_dst ^ netproto ^ transproto;
	}
	else
	{
		layer->_hash = FNV_BASIS ^ rotl(hash_dst, 32) ^ hash_src ^ netproto ^ transproto;
	}
	flow->layers_info |= NETWORK_LAYER_INFO;
}

static inline __u16 fill_vlan(struct __sk_buff *skb, int offset, struct flow *flow)
{
	struct link_layer *layer = &flow->link_layer;

	__u16 tci = load_half(skb, offset + offsetof(struct vlan, tci));
	__u16 protocol = load_half(skb, offset + offsetof(struct vlan, ethertype));
	__u16 vlanID = tci & 0x0fff;

	__u64 hash_vlan = 0;
	update_hash_half(&hash_vlan, vlanID);

	layer->_hash ^= hash_vlan;
	layer->id = (layer->id << 12) | vlanID;

	add_layer(flow, DOT1Q_LAYER);

	return protocol;
}

static inline void fill_vlans(struct __sk_buff *skb, __u16 *protocol, int *offset, struct flow *flow)
{
	if (*protocol == ETH_P_8021Q)
	{
#pragma unroll
		for (int i = 0; i < MAX_VLAN_LAYERS; i++)
		{
			*protocol = fill_vlan(skb, *offset, flow);
			*offset += 4;
			if (*protocol != ETH_P_8021Q)
			{
				break;
			}
		}
	}

	struct link_layer *layer = &flow->link_layer;
	if (skb->vlan_present)
	{
		__u16 vlanID = skb->vlan_tci & 0x0fff;
		layer->_hash ^= vlanID;
		layer->id = (layer->id << 12) | vlanID;

		add_layer(flow, DOT1Q_LAYER);
	}
}
