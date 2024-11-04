#include "common.bpf.h"
#include "traffic.bpf.h"

/*helper*/
static inline int udp_rcv_common(struct sk_buff *skb, bool is_fentry)
{
    if ((is_fentry ? !fentry_select : !kprobe_select) || !udp_info)
    {
        return 0;
    }
    return __udp_rcv(skb);
}
static inline int udp_enqueue_common(struct sock *sk, struct sk_buff *skb, bool is_fentry)
{
    if ((is_fentry ? !fentry_select : !kprobe_select) || !udp_info)
    {
        return 0;
    }
    return udp_enqueue_schedule_skb(sk, skb);
}
static inline int udp_send_common(struct sk_buff *skb, bool is_fentry)
{
    if ((is_fentry ? !fentry_select : !kprobe_select) || !udp_info)
    {
        return 0;
    }
    return __udp_send_skb(skb);
}
static inline int dev_hard_start_xmit_common(struct sk_buff *skb, bool is_fentry)
{
    if ((is_fentry ? !fentry_select : !kprobe_select) || !protocol_info)
    {
        return 0;
    }
    return __dev_hard_start_xmit(skb);
}

static inline int tcp_cleanup_common(struct sock *sk, int copied, bool is_fentry)
{
    if ((is_fentry ? !fentry_select : !kprobe_select) || !tcp_output_info)
    {
        return 0;
    }
    return __tcp_cleanup_rbuf(sk, copied);
}
static inline int tcp_send_common(struct sock *sk, struct msghdr *msg, size_t size, bool is_fentry)
{
    if ((is_fentry ? !fentry_select : !kprobe_select) || !tcp_output_info)
    {
        return 0;
    }
    return __tcp_sendmsg(sk, msg, size);
}
static inline int tcp_close_common(struct sock *sk, bool is_fentry)
{
    if ((is_fentry ? !fentry_select : !kprobe_select) || !tcp_output_info)
    {
        return 0;
    }
    return trace_tcp_close(sk);
}

static inline int eth_type_trans_common(struct sk_buff *skb, bool is_fentry)
{
    if ((is_fentry ? !fentry_select : !kprobe_select) || !protocol_info)
    {
        return 0;
    }
    return __eth_type_trans(skb);
}
static inline int tcp_connect_common(struct sock *sk, bool is_fentry)
{
    if ((is_fentry ? !fentry_select : !kprobe_select) || !packet_count)
    {
        return 0;
    }
    return __tcp_connect(sk);
}
static inline int tcp_rcv_state_process_common(struct sock *sk, struct sk_buff *skb, bool is_fentry)
{
    if ((is_fentry ? !fentry_select : !kprobe_select) || !packet_count)
    {
        return 0;
    }
    return __tcp_rcv_state_process(sk, skb);
}
static inline int tcp_send_fin_common(struct sock *sk, bool is_fentry)
{
    if ((is_fentry ? !fentry_select : !kprobe_select) || !packet_count)
    {
        return 0;
    }
    return __tcp_send_fin(sk);
}

// udp 
SEC("kprobe/udp_rcv")
int BPF_KPROBE(kp_udp_rcv, struct sk_buff *skb)
{
    return udp_rcv_common(skb, false);
}
SEC("fentry/udp_rcv")
int BPF_PROG(ft_udp_rcv, struct sk_buff *skb)
{
    return udp_rcv_common(skb, true);
}
SEC("kprobe/__udp_enqueue_schedule_skb")
int BPF_KPROBE(kp__udp_enqueue_schedule_skb, struct sock *sk, struct sk_buff *skb)
{
    return udp_enqueue_common(sk, skb, false);
}
SEC("fentry/__udp_enqueue_schedule_skb")
int BPF_PROG(ft__udp_enqueue_schedule_skb, struct sock *sk, struct sk_buff *skb)
{
    return udp_enqueue_common(sk, skb, true);
}
SEC("kprobe/udp_send_skb")
int BPF_KPROBE(kp_udp_send_skb, struct sk_buff *skb)
{
    return udp_send_common(skb, false);
}
SEC("fentry/udp_send_skb")
int BPF_PROG(ft_udp_send_skb, struct sk_buff *skb)
{
    return udp_send_common(skb, true);
}
SEC("kprobe/ip_send_skb")
int BPF_KPROBE(kp_ip_send_skb, struct net *net, struct sk_buff *skb)
{
    if (!kprobe_select || !udp_info)
    {
        return 0;
    }
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
int BPF_KPROBE(kp_tcp_sendmsg, struct sock *sk, struct msghdr *msg, size_t size)
{
    return tcp_send_common(sk, msg, size, false);
}

SEC("fentry/tcp_sendmsg")
int BPF_PROG(ft_tcp_sendmsg, struct sock *sk, struct msghdr *msg, size_t size)
{
    return tcp_send_common(sk, msg, size, true);
}
SEC("kprobe/tcp_cleanup_rbuf")
int BPF_KPROBE(kp_tcp_cleanup_rbuf, struct sock *sk, int copied)
{
    return tcp_cleanup_common(sk, copied, false);
}

SEC("fentry/tcp_cleanup_rbuf")
int BPF_PROG(ft_tcp_cleanup_rbuf, struct sock *sk, int copied)
{
    return tcp_cleanup_common(sk, copied, true);
}

SEC("kprobe/tcp_close")
int BPF_KPROBE(kp_tcp_close, struct sock *sk)
{
    return tcp_close_common(sk,false);
}

SEC("fentry/tcp_close")
int BPF_PROG(ft_tcp_close, struct sock *sk)
{
    return tcp_close_common(sk,true);
}

// count the usage of protocol ports
// receive
SEC("kprobe/eth_type_trans")
int BPF_KPROBE(kp_eth_type_trans, struct sk_buff *skb)
{
    return eth_type_trans_common(skb, false);
}
SEC("fentry/eth_type_trans")
int BPF_PROG(ft_eth_type_trans, struct sk_buff *skb)
{
    return eth_type_trans_common(skb, true);
}
SEC("kprobe/dev_hard_start_xmit")
int BPF_KPROBE(kp_dev_hard_start_xmit, struct sk_buff *skb)
{
    return dev_hard_start_xmit_common(skb, false);
}
SEC("fentry/dev_hard_start_xmit")
int BPF_PROG(ft_dev_hard_start_xmit, struct sk_buff *skb)
{
    return dev_hard_start_xmit_common(skb, true);
}
// iptables drop
SEC("kprobe/ipt_do_table")
int kp_ipt_do_table(struct pt_regs *ctx)
{
    return __ipt_do_table_start(ctx);
}

SEC("kretprobe/ipt_do_table")
int BPF_KRETPROBE(kp_ipt_do_table_ret)
{
    int ret = PT_REGS_RC(ctx);
    __ipt_do_table_ret(ctx, ret);
    return 0;
}
//skb drop
SEC("tracepoint/skb/kfree_skb")
int handle_kfree_skb(struct trace_event_raw_kfree_skb *ctx)
{
    return __kfree_skb(ctx);
}
//SYN、SYN-ACK、FIN
SEC("kprobe/tcp_connect")
int BPF_KPROBE(kp_tcp_connect, struct sock *sk)
{
    return tcp_connect_common(sk, false);
}

SEC("fentry/tcp_connect")
int BPF_PROG(ft_tcp_connect, struct sock *sk)
{
    return tcp_connect_common(sk, true);
}
SEC("kprobe/tcp_rcv_state_process")
int BPF_KPROBE(kp_tcp_rcv_state_process, struct sock *sk, struct sk_buff *skb)
{
    return tcp_rcv_state_process_common(sk, skb, false);
}
SEC("fentry/tcp_rcv_state_process")
int BPF_PROG(ft_tcp_rcv_state_process, struct sock *sk, struct sk_buff *skb)
{
    return tcp_rcv_state_process_common(sk, skb, true);
}
SEC("kprobe/tcp_send_fin")
int BPF_KPROBE(kp_tcp_send_fin, struct sock *sk)
{
    return tcp_send_fin_common(sk, false);
}
SEC("fentry/tcp_send_fin")
int BPF_PROG(ft_tcp_send_fin, struct sock *sk)
{
    return tcp_send_fin_common(sk, true);
}
//TCP CONN
SEC("tracepoint/tcp/tcp_rcv_space_adjust")
int handle_tcp_rcv_space_adjust(struct trace_event_raw_tcp_event_sk *ctx)
{
    return __tcp_rcv_space_adjust(ctx);
}

// SEC("tracepoint/tcp/tcp_send_reset")
// int handle_send_reset(struct trace_event_raw_tcp_event_sk_skb *ctx)
// {
// bpf_printk("ceshi11");
//    return 0;
// }
