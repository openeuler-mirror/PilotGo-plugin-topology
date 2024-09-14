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

// count the usage of protocol ports
// receive
SEC("kprobe/eth_type_trans")
int BPF_KPROBE(eth_type_trans, struct sk_buff *skb)
{
    return __eth_type_trans(skb);
}
// send
SEC("kprobe/dev_hard_start_xmit")
int BPF_KPROBE(dev_hard_start_xmit, struct sk_buff *skb)
{
    return __dev_hard_start_xmit(skb);
}
// iptables drop
SEC("kprobe/ipt_do_table")
int BPF_KPROBE(ipt_do_table, struct sk_buff *skb, u32 hook, struct nf_hook_state *state)
{
    struct xt_table *table = (struct xt_table *)PT_REGS_PARM4(ctx);
    return __ipt_do_table_start(ctx);
}
SEC("kretprobe/ipt_do_table")
int BPF_KRETPROBE(ipt_do_table_ret)
{
    int ret = PT_REGS_RC(ctx);
    return __ipt_do_table_ret(ctx, ret);
}
SEC("tracepoint/skb/kfree_skb")
int handle_kfree_skb(struct trace_event_raw_kfree_skb *ctx)
{
    return __kfree_skb(ctx);
}
// SYN-ACK total
//  SEC("kprobe/tcp_rcv_state_process ")
//  int BPF_KPROBE(tcp_rcv_state_process ,struct sock *sk,struct sk_buff *skb)
//  {
//      return __tcp_rcv_state_process(sk,skb);
//  }

SEC("kprobe/tcp_connect")
int BPF_KPROBE(tcp_connect, struct sock *sk)
{
    bpf_printk("tcp_connect");
    return __tcp_connect(sk);
}

SEC("kprobe/tcp_rcv_state_process")
int BPF_KPROBE(tcp_rcv_state_process, struct sock *sk, struct sk_buff *skb)
{
    bpf_printk("tcp_rcv_state_process");
    return __tcp_rcv_state_process(sk, skb);
}

SEC("kprobe/tcp_send_fin")
int BPF_KPROBE(tcp_send_fin, struct sock *sk)
{
    bpf_printk("tcp_send_fin");
    return __tcp_send_fin(sk);
}

// SEC("kprobe/tcp_send_reset")
// int handle_send_reset(struct pt_regs *ctx)
// {
//     return __tcp_send_reset(ctx);
// }

