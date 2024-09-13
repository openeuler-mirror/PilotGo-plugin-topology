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
// tcp
static __always_inline void handle_tcp_metrics(struct pt_regs *ctx, struct sock *sk, size_t size, bool is_tx, int pid)
{
    struct tcp_metrics_s *metrics = get_tcp_metrics(sk);
    if (!metrics)
    {
        return;
    }
    struct tcp_metrics_s tuple = {};
    get_tcp_tuple(sk, &tuple);
    metrics->pid = pid;

    if (is_tx)
    {
        metrics->client_ip = tuple.client_ip;
        metrics->server_ip = tuple.server_ip;
        metrics->client_port = tuple.client_port;
        metrics->server_port = tuple.server_port;
        metrics->tran_flag = 1;
        TCP_TX_DATA(metrics->tx_rx_stats, (int)size);
    }
    else
    {
        metrics->client_ip = tuple.server_ip;
        metrics->server_ip = tuple.client_ip;
        metrics->client_port = tuple.server_port;
        metrics->server_port = tuple.client_port;
        metrics->tran_flag = 0;
        get_tcp_tx_rx_segs(sk, &metrics->tx_rx_stats);
        TCP_RX_DATA(metrics->tx_rx_stats, size);
    }

    report_tx_rx(ctx, metrics, sk);
}
// send
static __always_inline int __tcp_sendmsg(struct pt_regs *ctx)
{
    struct sock *sk = (struct sock *)PT_REGS_PARM1(ctx);
    int pid = get_current_tgid();
    struct tcp_metrics_s tuple = {};

    get_tcp_tuple(sk, &tuple);
    size_t send_size = (size_t)PT_REGS_PARM3(ctx);
    // bpf_printk("Sending size: %zu\n", send_size);
    handle_tcp_metrics(ctx, sk, send_size, true, pid);
    return 0;
}

// recieve
static __always_inline int __tcp_cleanup_rbuf(struct pt_regs *ctx)
{
    struct sock *sk = (struct sock *)PT_REGS_PARM1(ctx);
    int pid = get_current_tgid();
    struct tcp_metrics_s tuple = {};
    get_tcp_tuple(sk, &tuple);
    int recieve_size = (int)PT_REGS_PARM2(ctx);
    if (recieve_size <= 0)
    {
        return 0;
    }
    // bpf_printk("recieve_size: %zu\n", recieve_size);
    handle_tcp_metrics(ctx, sk, (size_t)recieve_size, false, pid);
    return 0;
}

static __always_inline int process_packet(struct sk_buff *skb, bool is_tx)
{
    const struct ethhdr *eth = (struct ethhdr *)_R(skb, data);
    u16 protocol = _R(eth, h_proto);

    struct packet_info *pkt = bpf_ringbuf_reserve(&port_events_rb, sizeof(*pkt), 0);
    if (!pkt)
    {
        return 0;
    }

    if (_R(eth, h_proto) != __bpf_htons(ETH_P_IP))
    {
        bpf_ringbuf_discard(pkt, 0);
        return 0;
    }

    struct iphdr *ip = (struct iphdr *)(_R(skb, data) + 14);
    if (!ip)
    {
        bpf_ringbuf_discard(pkt, 0);
        return 0;
    }

    pkt->src_ip = _R(ip, saddr);
    pkt->dst_ip = _R(ip, daddr);
    pkt->proto = _R(ip, protocol);

    if (pkt->proto == IPPROTO_TCP)
    {
        struct tcphdr *tcp = (struct tcphdr *)(_R(skb, data) + sizeof(struct ethhdr) + sizeof(struct iphdr));
        pkt->src_port = _R(tcp, source);
        pkt->dst_port = _R(tcp, dest);
        pkt->proto = PROTO_TCP;
        // bpf_printk("TCP packet: src_port=%d, dst_port=%d\n", pkt->src_port, pkt->dst_port);
    }
    else if (pkt->proto == IPPROTO_UDP)
    {
        struct udphdr *udp = (struct udphdr *)(_R(skb, data) + sizeof(struct ethhdr) + sizeof(struct iphdr));
        pkt->src_port = _R(udp, source);
        pkt->dst_port = _R(udp, dest);
        pkt->proto = PROTO_UDP;
        // bpf_printk("UDP packet: src_port=%d, dst_port=%d\n", pkt->src_port, pkt->dst_port);
    }
    else if (pkt->proto == IPPROTO_ICMP)
    {
        pkt->proto = PROTO_ICMP;
        // bpf_printk("ICMP packet detected\n");
    }
    else
    {
        pkt->proto = PROTO_UNKNOWN;
        // bpf_printk("proto=%u\n", pkt->proto);
    }

    struct packet_count *count = count_packet(pkt->proto, is_tx);
    if (count)
    {
        pkt->count.tx_count = count->tx_count;
        pkt->count.rx_count = count->rx_count;
    }
    else
    {
        pkt->count.tx_count = 0;
        pkt->count.rx_count = 0;
    }

    bpf_ringbuf_submit(pkt, 0);

    return 0;
}

static __always_inline int __eth_type_trans(struct sk_buff *skb)
{
    return process_packet(skb, false); // receive
}

static __always_inline int __dev_hard_start_xmit(struct sk_buff *skb)
{
    return process_packet(skb, true); // send
}

static __always_inline int __ipt_do_table_start(struct pt_regs *ctx)
{

    u32 tid = bpf_get_current_pid_tgid();
    struct tid_map_value value = {};

    struct sk_buff *skb = (struct sk_buff *)PT_REGS_PARM1(ctx);
    struct nf_hook_state *state = (struct nf_hook_state *)PT_REGS_PARM3(ctx);
    struct xt_table *table = (struct xt_table *)PT_REGS_PARM4(ctx);
    u32 hook = (u32)PT_REGS_PARM2(ctx);

    value.skb = skb;
    value.state = state;
    value.hook = hook;
    value.table = table;

    bpf_map_update_elem(&inner_tid_map, &tid, &value, BPF_ANY);
    return 0;
}

static __always_inline int submit_event(struct pt_regs *ctx, struct tid_map_value *value, u32 drop_type)
{
    struct sock *sk;
    struct sk_buff *skb;
    struct drop_event *event;
    u64 addr;

    event = bpf_ringbuf_reserve(&perf_map, sizeof(struct drop_event), 0);
    if (!event)
    {
        return 0;
    }
    event->type = drop_type;
    skb = value->skb;

    bpf_probe_read(&sk, sizeof(sk), &skb->sk);
    // 栈
    event->kstack_sz =
        bpf_get_stack(ctx, event->kstack, sizeof(event->kstack), 0);
    event->pid = bpf_get_current_pid_tgid() >> 32;
    bpf_get_current_comm(&event->comm, sizeof(event->comm));
    fill_sk_skb(event, sk, skb);
    // 针对 iptables 的处理
    if (drop_type == DROP_IPTABLES_DROP)
    {
        struct xt_table *table = value->table;
        addr = bpf_core_xt_table_name(table);
        if (addr)
        {
            bpf_probe_read(event->name, sizeof(event->name), (void *)addr);
        }

        event->hook = value->hook;
    }
    bpf_ringbuf_submit(event, 0);
    return 1;
}

static __always_inline int handle_drop_event(struct pt_regs *ctx, int ret, struct tid_map_value *value, u32 drop_type)
{
    if (ret != NF_DROP || !value)
    {
        return 0;
    }
    return submit_event(ctx, value, drop_type);
}

static __always_inline int __ipt_do_table_ret(struct pt_regs *ctx, int ret)
{
    u32 tid = bpf_get_current_pid_tgid();
    struct tid_map_value *value = bpf_map_lookup_elem(&inner_tid_map, &tid);

    if (handle_drop_event(ctx, ret, value, DROP_IPTABLES_DROP))
    {
        bpf_map_delete_elem(&inner_tid_map, &tid);
    }
    return 0;
}

static __always_inline int __kfree_skb(struct trace_event_raw_kfree_skb *ctx)
{
    struct sk_buff *skb = ctx->skbaddr;
    if (!skb)
        return 0;
    struct iphdr *ip = extract_iphdr(skb);
    struct tcphdr *tcp = extract_tcphdr(skb);
    struct event devent = {0};
    get_tcp_pkt_tuple(&devent, ip, tcp);
    struct reasonissue *event;
    event = bpf_ringbuf_reserve(&trace_all_drop, sizeof(*event), 0);
    if (!event)
    {
        return 0;
    }
    event->client_ip = devent.client_ip;
    event->server_ip = devent.server_ip;
    event->client_port = devent.client_port;
    event->server_port = devent.server_port;
    event->pid = get_current_tgid();
    event->location = (long)ctx->location;
    event->protocol = ctx->protocol;
    // 丢包调用栈
    //  event->kstack_sz =
    //      bpf_get_stack(ctx, event->kstack, sizeof(event->kstack), 0);
    bpf_ringbuf_submit(event, 0);
    return 0;
}
