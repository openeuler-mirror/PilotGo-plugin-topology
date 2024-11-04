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
} port_events_rb SEC(".maps");
struct
{
    __uint(type, BPF_MAP_TYPE_RINGBUF);
    __uint(max_entries, 256 * 1024);
} trace_all_drop SEC(".maps");
struct
{
    __uint(type, BPF_MAP_TYPE_RINGBUF);
    __uint(max_entries, 1024);
} perf_map SEC(".maps");

struct
{
    __uint(type, BPF_MAP_TYPE_RINGBUF);
    __uint(max_entries, 256 * 1024);
} flags_rb SEC(".maps");

struct
{
    __uint(type, BPF_MAP_TYPE_RINGBUF);
    __uint(max_entries, 256 * 1024);
} rate_rb SEC(".maps");
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

struct
{
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, 65536);
    __type(key, u16);
    __type(value, struct packet_info);
} port_count SEC(".maps");

struct
{
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, 65536);
    __type(key, u32);
    __type(value, struct tid_map_value);
} inner_tid_map SEC(".maps");

// 用于存储 SYN 包计数
struct
{
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, 65536);
    __type(key, struct addr_pair);
    __type(value, u64);
} syn_count_map SEC(".maps");

// 用于存储 SYN-ACK 包计数
struct
{
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, 65536);
    __type(key, struct addr_pair);
    __type(value, u64);
} synack_count_map SEC(".maps");

// 用于存储 FIN 包计数
struct
{
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, 65536);
    __type(key, struct addr_pair);
    __type(value, u64);
} fin_count_map SEC(".maps");

struct
{
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, 65536);
    __type(key, struct sock *);
    __type(value, u64);
} tcp_rate_map SEC(".maps");

static int kprobe_select = 1, tcp_status_info = 1, fentry_select = 1, udp_info = 1, packet_count = 1, protocol_info = 1, tcp_output_info = 1;
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
    pkt_tuple->tran_flag = UDP;
}

static void get_tcp_pkt_tuple(void *pkt_tuple, struct iphdr *ip, struct tcphdr *tcp, int type)
{
    if (type == 1)
    { // struct tuple_key
        struct tuple_key *key = (struct tuple_key *)pkt_tuple;
        key->skbap.saddr = _R(ip, saddr);
        key->skbap.daddr = _R(ip, daddr);
        key->skbap.sport = __bpf_ntohs(_R(tcp, source));
        key->skbap.dport = __bpf_ntohs(_R(tcp, dest));
    }
    else if (type == 2)
    { // struct event
        struct event *event = (struct event *)pkt_tuple;
        event->client_ip = _R(ip, saddr);
        event->server_ip = _R(ip, daddr);
        event->client_port = __bpf_ntohs(_R(tcp, source));
        event->server_port = __bpf_ntohs(_R(tcp, dest));
        event->seq = __bpf_ntohl(_R(tcp, seq));
        event->ack = __bpf_ntohl(_R(tcp, ack_seq));
    }
}
static __always_inline void *bmloti(void *map, const void *key, const void *init)
{
    void *val;
    long err;

    val = bpf_map_lookup_elem(map, key);
    if (val)
        return val;

    err = bpf_map_update_elem(map, key, init, BPF_NOEXIST);
    if (err == 0)
    {
        return bpf_map_lookup_elem(map, key);
    }

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

static void get_tcp_tx_rx_segs(struct sock *sk, struct tcp_tx_rx *segs)
{
    struct tcp_sock *tcp_sk = (struct tcp_sock *)sk;
    segs->segs_in = _R(tcp_sk, segs_in);
    segs->segs_out = _R(tcp_sk, segs_out);
}

static __always_inline void report_tx_rx(struct sock_stats_s *sock_stats, struct sock *sk)
{
    sock_stats->metrics.report_flags |= TCP_PROBE_TXRX;
    void *buf = bpf_ringbuf_reserve(&tcp_output_rb, sizeof(struct tcp_metrics_s), 0);
    if (!buf)
    {
        return;
    }
    __builtin_memcpy(buf, &sock_stats->metrics, sizeof(struct tcp_metrics_s));
    bpf_ringbuf_submit(buf, 0);
}

static __always_inline struct tcp_metrics_s *get_tcp_metrics(struct sock *sk)
{
    struct sock_stats_s init_sock_stats = {0};
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
            return 0;
        }
        count = bpf_map_lookup_elem(&proto_stats, &proto);
        if (!count)
        {
            return 0;
        }
    }
    if (is_tx)
        __sync_fetch_and_add(&count->tx_count, 1);
    else
        __sync_fetch_and_add(&count->rx_count, 1);
    // bpf_printk("proto:%u count_tx:%llu count_rx:%llu\n", proto, count->tx_count, count->rx_count);
    return count;
}

static __always_inline u64 bpf_core_xt_table_name(void *ptr)
{
    struct xt_table *table = ptr;
    if (bpf_core_field_exists(table->name))
        return (u64)(&table->name[0]);
    return 0;
}
static __always_inline int fill_sk_skb(struct drop_event *event, struct sock *sk, struct sk_buff *skb)
{
    struct net *net = NULL;
    struct iphdr ih = {};
    struct tcphdr th = {};
    struct udphdr uh = {};
    u16 protocol = 0;
    bool has_netheader = false;
    u16 network_header, transport_header;
    char *head;
    event->has_sk = false;
    if (sk)
    {
        event->has_sk = true;
        bpf_probe_read(&event->skbap.daddr, sizeof(event->skbap.daddr), &sk->__sk_common.skc_daddr);
        bpf_probe_read(&event->skbap.dport, sizeof(event->skbap.dport), &sk->__sk_common.skc_dport);
        bpf_probe_read(&event->skbap.saddr, sizeof(event->skbap.saddr), &sk->__sk_common.skc_rcv_saddr);
        bpf_probe_read(&event->skbap.sport, sizeof(event->skbap.sport), &sk->__sk_common.skc_num);
        event->skbap.dport = bpf_ntohs(event->skbap.dport);
        protocol = _R(sk, __sk_common.skc_family);
        bpf_probe_read(&event->sk_state, sizeof(event->sk_state), (const void *)&sk->__sk_common.skc_state);
        //    bpf_printk(" IP:%U Dip:%u", event->skbap.saddr, event->skbap.daddr);
    }
   
    bpf_probe_read(&head, sizeof(head), &skb->head);
    bpf_probe_read(&network_header, sizeof(network_header), &skb->network_header);
    if (network_header != 0)
    {
        bpf_probe_read(&ih, sizeof(ih), head + network_header);
        has_netheader = true;
        event->skbap.saddr = ih.saddr;
        event->skbap.daddr = ih.daddr;
        event->skb_protocol = ih.protocol;
        transport_header = network_header + (ih.ihl << 2);
    }
    else
    {
        bpf_probe_read(&transport_header, sizeof(transport_header), &skb->transport_header);
    }
    switch (event->skb_protocol)
    {
    case IPPROTO_TCP:
        bpf_probe_read(&th, sizeof(th), head + transport_header);
        event->skbap.sport = bpf_ntohs(th.source);
        event->skbap.dport = bpf_ntohs(th.dest);
        break;
    case IPPROTO_UDP:
        if (transport_header != 0 && transport_header != 0xffff)
        {
            bpf_probe_read(&uh, sizeof(uh), head + transport_header);
            event->skbap.sport = bpf_ntohs(uh.source);
            event->skbap.dport = bpf_ntohs(uh.dest);
        }
        break;
    case IPPROTO_ICMP:
        break;
    default:
        return -1;
        break;
    }
    return 0;
}
static __always_inline void fill_tcp_packet_type(struct tuple_key *devent, struct sock *sk)
{
    devent->skbap.saddr = _R(sk, __sk_common.skc_rcv_saddr);
    devent->skbap.daddr = _R(sk, __sk_common.skc_daddr);
    devent->skbap.sport = _R(sk, __sk_common.skc_num);
    devent->skbap.dport = __bpf_ntohs(_R(sk, __sk_common.skc_dport));
}
#endif // __COMMON_BPF_H