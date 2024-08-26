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
