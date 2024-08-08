#ifndef __COMMON_BPF_H
#define __COMMON_BPF_H

#include "vmlinux.h"
#include "probe.h"
#include <bpf/bpf_core_read.h>
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>
#include <bpf/bpf_tracing.h>

char LICENSE[] SEC("license") = "Dual BSD/GPL";

/*rb helper*/
struct
{
    __uint(type, BPF_MAP_TYPE_RINGBUF);
    __uint(max_entries, 256 * 1024);
} udp_rb SEC(".maps");

struct
{
    __uint(type, BPF_MAP_TYPE_RINGBUF);
    __uint(max_entries, 256 * 1024);
} tcp_rb SEC(".maps");

/*map helper*/

struct
{
    __uint(type, BPF_MAP_TYPE_LRU_HASH);
    __uint(max_entries, MAX_COMM *MAX_PACKET);
    __type(key, struct event);
    __type(value, struct ktime_info);
} udp_map SEC(".maps");

struct
{
    __uint(type, BPF_MAP_TYPE_LRU_HASH);
    __uint(max_entries, MAX_COMM *MAX_PACKET);
    __type(key, struct sock *);
    __type(value, u64);
} tcp_status SEC(".maps");

/*funcation hepler*/
static __always_inline int get_current_tgid()
{
    return (int)(bpf_get_current_pid_tgid() >> PID);
}

static __always_inline struct tcphdr *extract_tcphdr(const struct sk_buff *skb)
{
    return (struct tcphdr *)((
        _R(skb, head) +
        _R(skb, transport_header)));
}

static __always_inline struct udphdr *extract_udphdr(const struct sk_buff *skb)
{
    return (struct udphdr *)(_R(skb, head) + _R(skb, transport_header));
}

static __always_inline struct iphdr *extract_iphdr(const struct sk_buff *skb)
{
    return (struct iphdr *)(_R(skb, head) + _R(skb, network_header));
}

static __always_inline void get_udp_pkt_tuple(struct event *pkt_tuple,
                                              struct iphdr *ip,
                                              struct udphdr *udp)
{
    pkt_tuple->client_ip = _R(ip, saddr);
    pkt_tuple->server_ip = _R(ip, daddr);
    pkt_tuple->client_port = __bpf_ntohs(_R(udp, source));
    pkt_tuple->server_port = __bpf_ntohs(_R(udp, dest));
    pkt_tuple->seq = 0;
    pkt_tuple->ack = 0;
    pkt_tuple->tran_flag = UDP;
}

static __always_inline void *bmloti(void *map, const void *key, const void *init)
{
    void *val;
    long err;
    val = bpf_map_lookup_elem(map, key);
    if (val)
        return val;
    err = bpf_map_update_elem(map, key, init,
                              BPF_NOEXIST);
    if (!err)
        return 0;

    return bpf_map_lookup_elem(map, key);
}

#endif // __COMMON_BPF_H