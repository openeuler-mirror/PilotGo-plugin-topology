#include "common.bpf.h"

// udp
static __always_inline struct event get_packet_tuple(struct sk_buff *skb)
{
    struct event pkt_tuple = {0};
    struct iphdr *ip = extract_iphdr(skb);
    struct udphdr *udp = extract_udphdr(skb);

    if (_R(ip, protocol) != IPPROTO_UDP)
    {
        return pkt_tuple;
    }
    get_udp_pkt_tuple(&pkt_tuple, ip, udp);
    return pkt_tuple;
}

static __always_inline struct ktime_info *loit(struct event *pkt_tuple)
{
    struct ktime_info zero = {0};
    return (struct ktime_info *)bmloti(&udp_map, pkt_tuple, &zero);
}

static __always_inline void submit_message(struct event *pkt_tuple, u64 tran_time, u8 rx, u16 len)
{
    u32 ptid = get_current_tgid();
    struct event *message = bpf_ringbuf_reserve(&udp_rb, sizeof(*message), 0);
    if (!message)
        return;

    message->tran_time = tran_time;
    message->client_ip = pkt_tuple->client_ip;
    message->server_ip = pkt_tuple->server_ip;
    message->client_port = pkt_tuple->client_port;
    message->server_port = pkt_tuple->server_port;
    message->udp_direction = rx;
    message->len = len;
    message->pid = ptid;
    bpf_get_current_comm(message->comm, sizeof(message->comm));
    bpf_ringbuf_submit(message, 0);
}

static __always_inline int __udp_rcv(struct sk_buff *skb)
{
    if (skb == NULL)
        return 0;
    struct event pkt_tuple = get_packet_tuple(skb);
    struct ktime_info *tinfo = loit(&pkt_tuple);
    if (!tinfo)
        return 0;

    tinfo->tran_time = NS_TIME();
    return 0;
}

static __always_inline int udp_enqueue_schedule_skb(struct sock *sk, struct sk_buff *skb)
{
    if (skb == NULL)
        return 0;
    struct event pkt_tuple = get_packet_tuple(skb);
    struct ktime_info *tinfo = bpf_map_lookup_elem(&udp_map, &pkt_tuple);
    if (!tinfo || tinfo->tran_time == 0)
        return 0;

    u64 tran_time = NS_TIME() - tinfo->tran_time;

    struct udphdr *udp = extract_udphdr(skb);

    submit_message(&pkt_tuple, tran_time, 1, __bpf_ntohs(_R(udp, len)));
    return 0;
}

static __always_inline int __udp_send_skb(struct sk_buff *skb)
{
    if (skb == NULL)
        return 0;

    struct event pkt_tuple = {0};
    struct sock *sk = _R(skb, sk);
    pkt_tuple.client_ip = _R(sk, __sk_common.skc_rcv_saddr);
    pkt_tuple.server_ip = _R(sk, __sk_common.skc_daddr);
    pkt_tuple.client_port = _R(sk, __sk_common.skc_num);
    pkt_tuple.server_port = __bpf_ntohs(_R(sk, __sk_common.skc_dport));
    pkt_tuple.tran_flag = UDP;

    struct ktime_info *tinfo = loit(&pkt_tuple);
    if (!tinfo)
        return 0;

    tinfo->tran_time = NS_TIME();
    return 0;
}

static __always_inline int __ip_send_skb(struct sk_buff *skb)
{
    if (skb == NULL)
        return 0;

    struct event pkt_tuple = get_packet_tuple(skb);

    struct ktime_info *tinfo = bpf_map_lookup_elem(&udp_map, &pkt_tuple);
    if (!tinfo || tinfo->tran_time == 0)
        return 0;

    struct udphdr *udp = extract_udphdr(skb);

    u64 tran_time = NS_TIME() - tinfo->tran_time;
    submit_message(&pkt_tuple, tran_time, 0, __bpf_ntohs(_R(udp, len)));
    return 0;
}

// tcp status
static __always_inline int __handle_tcp_state(struct trace_event_raw_inet_sock_set_state *ctx)
{
    u64 time, newtime;
    if (ctx->protocol != IPPROTO_TCP)
        return 0;

    struct sock *sk = (struct sock *)ctx->skaddr;
    u64 *before_time = bpf_map_lookup_elem(&tcp_status, &sk);
    newtime = NS_TIME();
    if (!before_time)
        time = 0;
    else
        time = newtime - *before_time;
    struct event tcpstate = {};
    tcpstate.oldstate = ctx->oldstate;
    tcpstate.newstate = ctx->newstate;
    tcpstate.family = ctx->family;
    tcpstate.client_port = ctx->sport;
    tcpstate.server_port = ctx->dport;
    bpf_probe_read_kernel(&tcpstate.client_ip, sizeof(tcpstate.client_ip),
                          &sk->__sk_common.skc_rcv_saddr);
    bpf_probe_read_kernel(&tcpstate.server_ip, sizeof(tcpstate.server_ip),
                          &sk->__sk_common.skc_daddr);
    tcpstate.tran_time = time;
    if (ctx->newstate == TCP_CLOSE)
        bpf_map_delete_elem(&tcp_status, &sk);
    else
        bpf_map_update_elem(&tcp_status, &sk, &newtime, BPF_ANY);

    struct event *message;
    message = bpf_ringbuf_reserve(&tcp_rb, sizeof(*message), 0);
    if (!message)
    {
        return 0;
    }
    message->pid = get_current_tgid();
    message->client_ip = tcpstate.client_ip;
    message->server_ip = tcpstate.server_ip;
    message->client_port = tcpstate.client_port;
    message->server_port = tcpstate.server_port;
    message->oldstate = tcpstate.oldstate;
    message->newstate = tcpstate.newstate;
    message->tran_time = tcpstate.tran_time;
    bpf_ringbuf_submit(message, 0);
    return 0;
}
