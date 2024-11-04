#ifndef __PROBE_H
#define __PROBE_H

typedef unsigned char u8;
typedef unsigned short u16;
typedef unsigned int u32;
typedef unsigned long long u64;
#define MTU_SIZE 1500
#define ETH_HLEN 14 // 以太网头部的长度
#define MAXSYMBOLS 300000
#define CACHEMAXSIZE 5
#define SK(sk) ((const struct sock *)(sk))
#define NS_TIME() (bpf_ktime_get_ns() / 1000)
#define _R(dst, src) BPF_CORE_READ(dst, src)
#define IPV6_LEN 16
#define MAX_COMM 16
#define MAX_PACKET 1000
#define BPF_MAP_TYPE_PERCPU_COUNTER 10
#define PID 32
#define MAX 256
#define MAX_SUM 1024
#define TASK_COMM_LEN 16
#define ETH_P_IP 0X0800
#define TCP 1
#define UDP 2
#define TIMEOUT_NS 5000ULL
#define TOP_N 5
#define MAX_ENTRIES 1000
#define NF_DROP 0
#define MAX_STACK_DEPTH 127
#define XT_TABLE_MAXNAMELEN 100
#define IPV4 2048
#define IPV6 34525
#define TCP_TX_DATA(data, delta) __sync_fetch_and_add(&((data).tx), (__u64)(delta))
#define TCP_RX_DATA(data, delta) __sync_fetch_and_add(&((data).rx), (__u64)(delta))
#define TCP_PROBE_TXRX (u32)(1 << 3)
#define RCV_SHUTDOWN 2
#define HASH_MAP_SIZE 1024
#define TIME_THRESHOLD_NS 200000000 // 200ms
typedef u64 stack_trace_t[MAX_STACK_DEPTH];

/* bpf.h struct helper */
struct addr_pair
{
    u32 saddr;
    u32 daddr;
    u16 sport;
    u16 dport;
};

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

struct reasonissue
{
    struct addr_pair skbap;
    long location;
    u16 protocol;
    int pid;
};
struct tcp_tx_rx
{
    size_t rx; // FROM tcp_cleanup_rbuf
    size_t tx; // FROM tcp_sendmsg
    u32 last_segs_out;
    u32 segs_out; // total number of segments sent
    u32 last_segs_in;
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
    bool is_reported;
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
    struct addr_pair skbap;
    u32 proto;
    int packet_count;
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

struct tid_map_value
{
    struct sk_buff *skb;
    struct nf_hook_state *state;
    struct xt_table *table;
    u32 hook;
    void *ctx;
};
struct drop_event
{
    u8 type;
    u8 name[32];
    u32 hook;
    u32 pid;
    u8 comm[16];
    u8 has_sk;
    u8 skb_protocol;
    u8 sk_state;
    u8 sk_protocol;
    struct addr_pair skbap;
    signed int kstack_sz;
    stack_trace_t kstack;
};

enum
{
    DROP_KFREE_SKB = 0,
    DROP_TCP_DROP,
    DROP_IPTABLES_DROP,
    DROP_NFCONNTRACK_DROP,
    UNK,
};

static const char *drop_type_str[] = {
    "DROP_KFREE_SKB",
    "DROP_TCP_DROP",
    "DROP_IPTABLES_DROP",
    "DROP_NFCONNTRACK_DROP",
    "UNKNOWN"};

static const char *protocol_names[] = {
    [IPPROTO_ICMP] = "ICMP", // 1
    [IPPROTO_TCP] = "TCP",   // 6
    [IPPROTO_UDP] = "UDP",   // 17
};

struct SymbolEntry
{
    unsigned long addr;
    char name[30];
};
// 4-tuple 结构体定义
struct tuple_key
{
    struct addr_pair skbap;
    u8 packet_type;
};
struct packet_stats
{
    u64 syn_count;
    u64 synack_count;
    u64 fin_count;
    struct tuple_key key;
};
struct tcp_event
{
    struct addr_pair skbap;
    struct packet_stats sum;
};
struct tcp_rate
{
    struct addr_pair skbap;
    u64 tcp_ato;
    u64 tcp_rto;
    u64 tcp_delack_max;
    u32 pid;
};

#endif
