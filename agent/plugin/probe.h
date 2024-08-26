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
    char comm[TASK_COMM_LEN];
};

#endif
