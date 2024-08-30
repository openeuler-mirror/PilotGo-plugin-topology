#ifndef __PROBE_H
#define __PROBE_H

typedef unsigned char u8;
typedef unsigned short u16;
typedef unsigned int u32;
typedef unsigned long long u64;

#define SK(sk) ((const struct sock *)(sk))
#define NS_TIME() (bpf_ktime_get_ns() / 1000)
#define _R(dst, src) BPF_CORE_READ(dst, src)
#define IPV6_LEN 16
#define MAX_COMM 16
#define MAX_PACKET 1000
#define AF_INET 2
#define BPF_MAP_TYPE_PERCPU_COUNTER 10
#define PID 32
#define MAX 256
#define MAX_SUM 1024
#define TASK_COMM_LEN 16
#define ETH_P_IP 0X0800
#define IPPROTO_TCP 6
#define IPPROTO_UDP 17
#define TCP 1
#define UDP 2
#define TIMEOUT_NS 5000ULL

#define TCP_TX_DATA(data, delta) __sync_fetch_and_add(&((data).tx), (__u64)(delta))
#define TCP_RX_DATA(data, delta) __sync_fetch_and_add(&((data).rx), (__u64)(delta))
#define TCP_PROBE_TXRX (u32)(1 << 3)

/* bpf.h struct helper */
struct ktime_info
{
    u64 qdisc_time;
    u64 mac_time;
    u64 ip_time;
    u64 tcp_time;
    u64 tran_time;
    u64 app_time;
};

struct event
{
    u32 client_ip;
    u32 server_ip;
    u16 client_port;
    u16 server_port;
    u32 seq;
    u32 ack;
    u32 tran_flag;
    u32 len;
    int pid;
    int udp_direction;
    u16 protocol;
    int tran_time;
    u16 family;
    char comm[TASK_COMM_LEN];
    int oldstate;
    int newstate;
    u8 type;
};

struct tcp_tx_rx
{
    u64 rx; // FROM tcp_cleanup_rbuf
    u64 tx; // FROM tcp_sendmsg
    u32 last_time_segs_out;
    u32 segs_out; // total number of segments sent
    u32 last_time_segs_in;
    u32 segs_in; // total number of segments in
};

struct tcp_metrics_s
{
    int pid;
    u32 client_ip;
    u32 server_ip;
    u16 client_port;
    u16 server_port;
    u32 report_flags;
    u32 tran_flag;
    struct tcp_tx_rx tx_rx_stats;
};

struct sock_stats_s
{
    u64 txrx_ts;
    struct tcp_metrics_s metrics;
};

enum
{
    PROTO_TCP = 0,
    PROTO_UDP,
    PROTO_ICMP,
    PROTO_UNKNOWN,
    PROTO_MAX,
};

struct packet_count
{
    u64 rx_count; 
    u64 tx_count; 
};

struct packet_info
{
    __u32 src_ip;              
    __u32 dst_ip;              
    __u16 src_port;            
    __u16 dst_port;            
    __u32 proto;              
    struct packet_count count; 
};

struct protocol_stats
{
    uint64_t rx_count;
    uint64_t tx_count;
};

static const char *tcp_states[] = {
    [1] = "ESTABLISHED",
    [2] = "SYN_SENT",
    [3] = "SYN_RECV",
    [4] = "FIN_WAIT1",
    [5] = "FIN_WAIT2",
    [6] = "TIME_WAIT",
    [7] = "CLOSE",
    [8] = "CLOSE_WAIT",
    [9] = "LAST_ACK",
    [10] = "LISTEN",
    [11] = "CLOSING",
    [12] = "NEW_SYN_RECV",
    [13] = "UNKNOWN",
};

static const char *protocol[] = {
    [0] = "TCP",
    [1] = "UDP",
    [2] = "ICMP",
    [3] = "UNKNOWN",
};
#endif
