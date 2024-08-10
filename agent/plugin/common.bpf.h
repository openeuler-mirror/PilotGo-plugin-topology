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

struct
{
    __uint(type, BPF_MAP_TYPE_RINGBUF);
    __uint(max_entries, 256 * 1024);
} tcp_output_rb SEC(".maps");

struct
{
    __uint(type, BPF_MAP_TYPE_RINGBUF);
    __uint(max_entries, 256 * 1024);
} port_events SEC(".maps");

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

struct
{
    __uint(type, BPF_MAP_TYPE_LRU_HASH);
    __uint(max_entries, MAX_COMM *MAX_PACKET);
    __type(key, struct sock *);
    __type(value, struct sock_stats_s);
} tcp_link_map SEC(".maps");
// packets
struct
{
    __uint(type, BPF_MAP_TYPE_ARRAY);
    __uint(max_entries, MAX_COMM *MAX_PACKET);
    __type(key, u32);
    __type(value, struct packet_count);
} proto_stats SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, 65536);  
    __type(key, u16);  
    __type(value, struct packet_info);
} port_count SEC(".maps");

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
static __always_inline char is_period_txrx(struct sock *sk)
{
    struct sock_stats_s *sock_stats = bpf_map_lookup_elem(&tcp_link_map, &sk);
    if (!sock_stats)
    {
        bpf_printk("1111: No sock_stats found for socket.");
        return 0;
    }

    u64 current_time = NS_TIME();
    u64 last_time = sock_stats->txrx_ts;
    u64 elapsed_time = current_time - last_time;

    //  bpf_printk("Current time: %llu, Last time: %llu, Elapsed time: %llu", current_time, last_time, elapsed_time);

    if ((current_time > last_time) && (elapsed_time >= TIMEOUT_NS))
    {
        sock_stats->txrx_ts = current_time;
        return 1;
    }
    return 0;
}

static __always_inline void report_tx_rx(void *ctx, struct tcp_metrics_s *metrics, struct sock *sk)
{

    if (!is_period_txrx(sk))
    {
        return;
    }

    u32 last_time_segs_out = metrics->tx_rx_stats.segs_out; // save the number of last sent segments
    u32 last_time_segs_in = metrics->tx_rx_stats.segs_in;   // recieve segments
    metrics->report_flags |= TCP_PROBE_TXRX;
    void *buf = bpf_ringbuf_reserve(&tcp_output_rb, sizeof(struct tcp_metrics_s), 0);
    if (!buf)
    {
        return;
    }

    __builtin_memcpy(buf, metrics, sizeof(struct tcp_metrics_s));
    metrics->report_flags &= ~TCP_PROBE_TXRX;
    __builtin_memset(&(metrics->tx_rx_stats), 0x0, sizeof(metrics->tx_rx_stats));

    metrics->tx_rx_stats.last_time_segs_in = last_time_segs_in;
    metrics->tx_rx_stats.last_time_segs_out = last_time_segs_out;
    // bpf_printk("=========Reporting TX/RX. Last segs_out: %u, Last segs_in: %u", metrics->tx_rx_stats.last_time_segs_out, metrics->tx_rx_stats.last_time_segs_in);

    bpf_ringbuf_submit(buf, 0);
}

static void get_tcp_tx_rx_segs(struct sock *sk, struct tcp_tx_rx *segs)
{
    struct tcp_sock *tcp_sk = (struct tcp_sock *)sk;
    segs->segs_in = _R(tcp_sk, segs_in);
    segs->segs_out = _R(tcp_sk, segs_out);
    // bpf_printk("Got TCP TX/RX segs. segs_in: %u, segs_out: %u", segs->segs_in, segs->segs_out);
}

static __always_inline struct tcp_metrics_s *get_tcp_metrics(struct sock *sk)
{
    struct sock_stats_s init_sock_stats = {};
    struct sock_stats_s *sock_stats;

    sock_stats = (struct sock_stats_s *)bmloti(&tcp_link_map, &sk, &init_sock_stats);

    if (!sock_stats)
    {
        return NULL;
    }

    return &(sock_stats->metrics);
}

static __always_inline void get_tcp_tuple(struct sock *sk, struct tcp_metrics_s *tuple)
{
    tuple->client_ip = _R(sk, __sk_common.skc_rcv_saddr);
    tuple->server_ip = _R(sk, __sk_common.skc_daddr);
    tuple->client_port = _R(sk, __sk_common.skc_num);
    tuple->server_port = __bpf_ntohs(_R(sk, __sk_common.skc_dport));
}


static __always_inline struct packet_count *count_packet(__u32 proto, bool is_tx)
{
    struct packet_count *count;
    struct packet_count initial_count = {0};

    count = bpf_map_lookup_elem(&proto_stats, &proto);
    if (!count)
    {
        initial_count.tx_count = 0;
        initial_count.rx_count = 0;
        if (bpf_map_update_elem(&proto_stats, &proto, &initial_count, BPF_ANY))
        {
            // bpf_printk("proto:%u failed to initialize count\n", proto);
            return NULL;
        }
        count = bpf_map_lookup_elem(&proto_stats, &proto);
        if (!count)
        {
            // bpf_printk("proto:%u count is NULL after initialization\n", proto);
            return NULL;
        }
    }

    if (is_tx)
        __sync_fetch_and_add(&count->tx_count, 1);
    else
        __sync_fetch_and_add(&count->rx_count, 1);

    // bpf_printk("proto:%u count_tx:%llu count_rx:%llu\n", proto, count->tx_count, count->rx_count);
    return count;
}


#endif // __COMMON_BPF_H