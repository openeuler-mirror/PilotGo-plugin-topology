#include "common.bpf.h"
#include "traffic.bpf.h"

// udp recieve
SEC("kprobe/udp_rcv")
int BPF_KPROBE(udp_rcv, struct sk_buff *skb)
{
    return __udp_rcv(skb);
}

SEC("kprobe/__udp_enqueue_schedule_skb")
int BPF_KPROBE(__udp_enqueue_schedule_skb, struct sock *sk,
               struct sk_buff *skb)
{
    return udp_enqueue_schedule_skb(sk, skb);
}

// send
SEC("kprobe/udp_send_skb")
int BPF_KPROBE(udp_send_skb, struct sk_buff *skb)
{
    return __udp_send_skb(skb);
}

SEC("kprobe/ip_send_skb")
int BPF_KPROBE(ip_send_skb, struct net *net, struct sk_buff *skb)
{

    return __ip_send_skb(skb);
}

// tcp status
SEC("tracepoint/sock/inet_sock_set_state")
int handle_tcp_state(struct trace_event_raw_inet_sock_set_state *ctx)
{
    return __handle_tcp_state(ctx);
}

// protocol  recieve
SEC("kprobe/tcp_sendmsg")
int __handle_tcp_sendmsg(struct pt_regs *ctx)
{
    return __tcp_sendmsg(ctx);
}

SEC("kprobe/tcp_cleanup_rbuf")
int __handle_tcp_cleanup_rbuf(struct pt_regs *ctx)
{
    return __tcp_cleanup_rbuf(ctx);
}
