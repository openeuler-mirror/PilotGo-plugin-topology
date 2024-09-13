#include <argp.h>
#include <arpa/inet.h>
#include <bpf/bpf.h>
#include <bpf/libbpf.h>
#include <netinet/in.h>
#include <netinet/tcp.h>
#include <signal.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include <sys/time.h>
#include <unistd.h>
#include "probe.h"
#include "probe.skel.h"
#include "probe.h"
int stack_map_fd; // 在全局范围内声明
static volatile bool exiting = false;
static int udp_info = 0, tcp_status_info = 0, tcp_output_info = 0, protocol_info = 0, port_distribution = 0, drop_info = 0, drop_skb = 0;
struct protocol_stats proto_stats[256] = {0};
static int interval = 5, entry_count = 0; // 每5 秒计算一次
struct packet_info entries[MAX_ENTRIES];
time_t start_time = 0;

const char argp_program_doc[] = "Trace time delay in network subsystem \n";

static const struct argp_option opts[] = {
    {"udp", 'u', 0, 0, "trace the udp message"},
    {"tcp_status_info", 't', 0, 0, "trace the tcp states"},
    {"tcp_output_info", 'o', 0, 0, "trace the tcp flow"},
    {"protocol_info", 'p', 0, 0, "statistics on the use of different protocols"},
    {"port_distribution_info", 'P', 0, 0, "statistical use of top10 destination ports"},
    {"drop_info", 'i', 0, 0, "trace the iptables drop"},
    {"drop_skb", 'd', 0, 0, "trace the all skb drop"},
    {},
};

static error_t parse_arg(int key, char *arg, struct argp_state *state)
{
    switch (key)
    {
    case 'u':
        udp_info = 1;
        break;
    case 't':
        tcp_status_info = 1;
        break;
    case 'o':
        tcp_output_info = 1;
        break;
    case 'p':
        protocol_info = 1;
        break;
    case 'P':
        port_distribution = 1;
        break;
    case 'i':
        drop_info = 1;
        break;
    case 'd':
        drop_skb = 1;
        break;
    default:
        return ARGP_ERR_UNKNOWN;
    }
    return 0;
}

static const struct argp argp = {
    .options = opts,
    .parser = parse_arg,
    .doc = argp_program_doc,
};

static void sig_handler(int sig)
{
    exiting = true;
}

static int libbpf_print_fn(enum libbpf_print_level level, const char *format, va_list args)
{
    return vfprintf(stderr, format, args);
}

static void format_ip_address(__be32 ip, char *buffer, size_t buffer_size)
{
    inet_ntop(AF_INET, &ip, buffer, buffer_size);
}

static int print_udp_event_info(void *ctx, void *packet_info, size_t data_sz)
{
    if (!udp_info)
    {
        return 0;
    }
    char s_str[INET_ADDRSTRLEN];
    char d_str[INET_ADDRSTRLEN];
    const struct event *pack_info = (const struct event *)packet_info;
    if (pack_info->client_port == 0 || pack_info->server_port == 0 || pack_info->pid == 0)
    {
        return 0;
    }

    format_ip_address(pack_info->client_ip, s_str, sizeof(s_str));
    format_ip_address(pack_info->server_ip, d_str, sizeof(d_str));

    printf("%-20d %-20s %-20s %-20u %-20d %-20s %-20d %-20d %-20d\n",
           pack_info->pid,
           s_str,
           d_str,
           pack_info->client_port,
           pack_info->server_port,
           pack_info->comm,
           pack_info->tran_time,
           pack_info->udp_direction,
           pack_info->len);
    return 0;
}

static int print_tcp_state_info(void *ctx, void *packet_info, size_t data_sz)
{
    if (!tcp_status_info)
    {
        return 0;
    }
    const struct event *pack_info = (const struct event *)packet_info;
    char s_str[INET_ADDRSTRLEN];
    char d_str[INET_ADDRSTRLEN];

    format_ip_address(pack_info->client_ip, s_str, sizeof(s_str));
    format_ip_address(pack_info->server_ip, d_str, sizeof(d_str));
    if (pack_info->client_ip == 0 || pack_info->server_ip == 0)
    {
        return 0;
    }
    printf("%-20d %-20s %-20s %-20d %-20d %-20s %-20s %-20d\n",
           pack_info->pid,
           s_str,
           d_str,
           pack_info->client_port,
           pack_info->server_port,
           tcp_states[pack_info->oldstate],
           tcp_states[pack_info->newstate],
           pack_info->tran_time);
    return 0;
}

static int print_tcp_flow_info(void *ctx, void *packet_info, size_t data_sz)
{
    if (!tcp_output_info)
    {
        return 0;
    }
    const struct tcp_metrics_s *pack_info = (const struct tcp_metrics_s *)packet_info;
    char s_str[INET_ADDRSTRLEN];
    char d_str[INET_ADDRSTRLEN];
    if (pack_info->client_ip > 0xFFFFFFFF || pack_info->server_ip > 0xFFFFFFFF || pack_info->pid <= 0)
    {
        return 0;
    }

    format_ip_address(pack_info->client_ip, s_str, sizeof(s_str));
    format_ip_address(pack_info->server_ip, d_str, sizeof(d_str));

    printf("%-20d %-20s %-20s %-20d %-20d %-20llu %-20llu %-20u %-20u %-20d\n",
           pack_info->pid,
           s_str,
           d_str,
           pack_info->client_port,
           pack_info->server_port,
           pack_info->tx_rx_stats.rx,
           pack_info->tx_rx_stats.tx,
           pack_info->tx_rx_stats.segs_in,
           pack_info->tx_rx_stats.segs_out,
           pack_info->tran_flag);
    return 0;
}

// function for calculating and printing the proportion of protocols
void calculate_protocol_usage(struct protocol_stats proto_stats[], int num_protocols, int interval)
{
    static uint64_t last_rx[256] = {0}, last_tx[256] = {0};
    uint64_t current_rx = 0, current_tx = 0;
    uint64_t delta_rx[256] = {0}, delta_tx[256] = {0};

    for (int i = 0; i < num_protocols; i++)
    {
        if (proto_stats[i].rx_count >= last_rx[i])
        {
            delta_rx[i] = proto_stats[i].rx_count - last_rx[i];
        }
        else
        {
            delta_rx[i] = proto_stats[i].rx_count;
        }

        if (proto_stats[i].tx_count >= last_tx[i])
        {
            delta_tx[i] = proto_stats[i].tx_count - last_tx[i];
        }
        else
        {
            delta_tx[i] = proto_stats[i].tx_count;
        }

        current_rx += delta_rx[i];
        current_tx += delta_tx[i];

        last_rx[i] = proto_stats[i].rx_count;
        last_tx[i] = proto_stats[i].tx_count;
    }

    // Proportion of agreements
    printf("===============================Protocol Usage in Last %d Seconds:\n", interval);
    printf("Total_rx_count:%ld Total_tx_count:%ld\n", current_rx, current_tx);

    if (current_rx > 0)
    {
        printf("Receive Protocol Usage:\n");
        for (int i = 0; i < num_protocols; i++)
        {
            if (delta_rx[i] > 0)
            {
                double rx_percentage = (double)delta_rx[i] / current_rx * 100;
                printf("Protocol %s: %.2f%% Rx_count:%ld\n", protocol[i], rx_percentage, delta_rx[i]);
            }
        }
    }

    if (current_tx > 0)
    {
        printf("Transmit Protocol Usage:\n");
        for (int i = 0; i < num_protocols; i++)
        {
            if (delta_tx[i] > 0)
            {
                double tx_percentage = (double)delta_tx[i] / current_tx * 100;
                printf("Protocol %s: %.2f%% Tx_count:%ld\n", protocol[i], tx_percentage, delta_tx[i]);
            }
        }
    }

    memset(proto_stats, 0, num_protocols * sizeof(struct protocol_stats));
}

int compare_by_pps(const void *a, const void *b)
{
    return ((struct packet_info *)b)->packet_count - ((struct packet_info *)a)->packet_count;
}

void init_start_time()
{
    start_time = time(NULL);
}

int find_port_entry(int dst_port, int proto)
{
    for (int i = 0; i < entry_count; i++)
    {
        if (entries[i].dst_port == dst_port && entries[i].proto == proto)
        {
            return i;
        }
    }
    return -1;
}

static int print_drop(void *ctx, void *packet_info, size_t data_sz)
{
    if (!drop_info)
    {
        return 0;
    }
    const struct drop_event *event = (const struct drop_event *)packet_info;

    char s_str[INET_ADDRSTRLEN];
    char d_str[INET_ADDRSTRLEN];

    if (!drop_info)
    {
        return 0;
    }

    format_ip_address(event->skbap.saddr, s_str, sizeof(s_str));
    format_ip_address(event->skbap.daddr, d_str, sizeof(d_str));
    const char *type_str = event->type >= 0 && event->type < 4 ? drop_type_str[event->type] : "UNKNOWN";
    // protocol string
    const char *proto_str = (event->skb_protocol >= 0 && event->skb_protocol < sizeof(protocol_names) / sizeof(char *) &&
                             protocol_names[event->skb_protocol])
                                ? protocol_names[event->skb_protocol]
                                : "UNKNOWN";
    printf("%-20d %-20s %-20s %-20d %-20d %-20s %-20s\n",
           event->pid, s_str, d_str, event->skbap.sport, event->skbap.dport,
           proto_str, type_str);
    return 0;
}
static int print_drop_skb(void *ctx, void *packet_info, size_t data_sz)
{
    if (!drop_skb)
    {
        return 0;
    }
    const struct event *event = (const struct event *)packet_info;
    char s_str[INET_ADDRSTRLEN];
    char d_str[INET_ADDRSTRLEN];
    char protol[6];
    if (event->client_ip == 0 && event->server_ip == 0)
    {
        return 0;
    }
    format_ip_address(event->client_ip, s_str, sizeof(s_str));
    format_ip_address(event->server_ip, d_str, sizeof(d_str));
    if (event->protocol == IPV4)
    {
        strcpy(protol, "ipv4");
    }
    else if (event->protocol == IPV6)
    {
        strcpy(protol, "ipv6");
    }
    else
    {
        strcpy(protol, "other");
    }
    printf("%-20d %-20s %-20s %-20d %-20d %-20s %-34lx \n",
           event->pid, s_str, d_str, event->client_port, event->server_port, protol, event->location);
    return 0;
}
static int print_count_protocol_use(void *ctx, void *packet_info, size_t data_sz)
{
    const struct packet_info *pack_protocol_info = (const struct packet_info *)packet_info;

    if (protocol_info)
    {
        proto_stats[pack_protocol_info->proto].rx_count = pack_protocol_info->count.rx_count;
        proto_stats[pack_protocol_info->proto].tx_count = pack_protocol_info->count.tx_count;
    }
    if (port_distribution)
    {
        // 查找当前端口号和协议号是否已经存在于 entries 数组中
        int index = find_port_entry(pack_protocol_info->dst_port, pack_protocol_info->proto);
        if (index != -1)
        {
            entries[index].packet_count++;
        }
        else
        {
            if (entry_count >= MAX_ENTRIES)
            {
                printf("entry_count big");
                return 0;
            }
            entries[entry_count].dst_port = pack_protocol_info->dst_port;
            entries[entry_count].proto = pack_protocol_info->proto;
            entries[entry_count].packet_count = 1;
            entry_count++;
        }
    }
    return 0;
}
static int print_top_5_keys()
{
    printf("Entry count: %d\n", entry_count);
    // 使用 qsort 对 PPS 进行排序
    qsort(entries, entry_count, sizeof(struct packet_info), compare_by_pps);

    // 输出前10个最频繁使用的端口号及其 PPS 值和协议号
    printf("==========Top %d Ports by PPS:\n", TOP_N);
    for (int i = 0; i < TOP_N && i < entry_count; i++)
    {
        const char *proto_str = (entries[i].proto >= 0 && entries[i].proto <= 3) ? protocol[entries[i].proto] : "UNKNOWN";
        printf("Port: %d, PPS: %d, Protocol: %s\n", entries[i].dst_port, entries[i].packet_count, proto_str);
    }
    memset(entries, 0, entry_count * sizeof(struct packet_info));
    entry_count = 0;
    return 0;
}
int main(int argc, char **argv)
{

    struct probe_bpf *skel;
    int err = 0;
    struct ring_buffer *udp_rb = NULL;
    struct ring_buffer *tcp_rb = NULL;
    struct ring_buffer *tcp_output_rb = NULL;
    struct ring_buffer *port_events_rb = NULL;
    struct ring_buffer *perf_map = NULL;
    struct ring_buffer *trace_all_drop = NULL;
    /* Parse command line arguments */
    err = argp_parse(&argp, argc, argv, 0, NULL, NULL);
    if (err)
        return err;

    libbpf_set_strict_mode(LIBBPF_STRICT_ALL);
    /* Set up libbpf errors and debug info callback */
    libbpf_set_print(libbpf_print_fn);

    /* Cleaner handling of Ctrl-C */
    signal(SIGINT, sig_handler);
    signal(SIGTERM, sig_handler);

    /* Load and verify BPF application */
    skel = probe_bpf__open();
    if (!skel)
    {
        fprintf(stderr, "Failed to open and load BPF skeleton\n");
        return 1;
    }

    /* Load & verify BPF programs */
    err = probe_bpf__load(skel);
    if (err)
    {
        fprintf(stderr, "Failed to load and verify BPF skeleton\n");
        goto cleanup;
    }
    /* Attach tracepoints */
    err = probe_bpf__attach(skel);
    if (err)
    {
        fprintf(stderr, "Failed to attach BPF skeleton\n");
        goto cleanup;
    }

    /* Set up ring buffer polling */
    udp_rb = ring_buffer__new(bpf_map__fd(skel->maps.udp_rb), print_udp_event_info, NULL, NULL);
    if (!udp_rb)
    {
        err = -1;
        fprintf(stderr, "Failed to create ring buffer\n");
        goto cleanup;
    }
    tcp_rb = ring_buffer__new(bpf_map__fd(skel->maps.tcp_rb), print_tcp_state_info, NULL, NULL);
    if (!tcp_rb)
    {
        err = -1;
        fprintf(stderr, "Failed to create ring buffer\n");
        goto cleanup;
    }
    tcp_output_rb = ring_buffer__new(bpf_map__fd(skel->maps.tcp_output_rb), print_tcp_flow_info, NULL, NULL);
    if (!tcp_output_rb)
    {
        err = -1;
        fprintf(stderr, "Failed to create ring buffer\n");
        goto cleanup;
    }
    port_events_rb = ring_buffer__new(bpf_map__fd(skel->maps.port_events_rb), print_count_protocol_use, NULL, NULL);
    if (!port_events_rb)
    {
        err = -1;
        fprintf(stderr, "Failed to create ring buffer\n");
        goto cleanup;
    }
    perf_map = ring_buffer__new(bpf_map__fd(skel->maps.perf_map), print_drop, NULL, NULL);
    if (!perf_map)
    {
        err = -1;
        fprintf(stderr, "Failed to create ring buffer\n");
        goto cleanup;
    }
    trace_all_drop = ring_buffer__new(bpf_map__fd(skel->maps.trace_all_drop), print_drop_skb, NULL, NULL);
    if (!trace_all_drop)
    {
        err = -1;
        fprintf(stderr, "Failed to create ring buffer\n");
        goto cleanup;
    }

    /* Process events */
    if (udp_info)
    {
        printf("%-20s %-20s %-20s %-20s %-20s %-20s %-20s %-20s %-20s\n", "Pid", "Client_ip", "Server_ip", "Client_port", "Server_port", "Comm", "Tran_time/μs", "Direction", "len/byte");
    }
    if (tcp_status_info)
    {
        printf("%-20s %-20s %-20s %-20s %-20s %-20s %-20s %-20s \n", "Pid", "Client_ip", "Server_ip", "Client_port", "Server_port", "oldstate", "newstate", "time/μs");
    }
    if (tcp_output_info)
    {
        printf("%-20s %-20s %-20s %-20s %-20s %-20s %-20s %-20s %-20s %-20s\n", "Pid", "Client_ip", "Server_ip", "Client_port", "Server_port", "Send/bytes", "receive/bytes", "segs_in", "segs_out", "Direction");
    }
    if (drop_info)
    {
        printf("%-20s %-20s %-20s %-20s %-20s %-20s %-20s \n", "Pid", "Client_ip", "Server_ip", "Client_port", "Server_port", "Protocol", "Drop_type");
    }
    if (drop_skb)
    {
        printf("%-20s %-20s %-20s %-20s %-20s %-20s %-20s \n", "Pid", "Client_ip", "Server_ip", "Client_port", "Server_port", "Protocol", "DROP_addr");
    }
    if (protocol_info)
    {
        printf("==========Proportion of each agreement==========\n");
    }
    if (port_distribution)
    {
        printf("==========port_distribution==========\n");
    }
    start_time = time(NULL);
    struct timeval start, end;
    gettimeofday(&start, NULL);
    while (!exiting)
    {
        err = ring_buffer__poll(udp_rb, 100 /* timeout, ms */);
        err = ring_buffer__poll(tcp_rb, 100 /* timeout, ms */);
        err = ring_buffer__poll(tcp_output_rb, 100 /* timeout, ms */);
        err = ring_buffer__poll(port_events_rb, 100 /* timeout, ms */);
        err = ring_buffer__poll(perf_map, 100 /* timeout, ms */);
        err = ring_buffer__poll(trace_all_drop, 100 /* timeout, ms */);
        /* Ctrl-C will cause -EINTR */
        // Regularly calculate and print the proportion of agreements
        if (protocol_info)
        {
            if (time(NULL) - start_time >= interval)
            {
                calculate_protocol_usage(proto_stats, 256, interval);
                start_time = time(NULL); // reset time
            }
        }

        if (err == -EINTR)
        {
            err = 0;
            break;
        }
        if (err < 0)
        {
            printf("Error polling perf buffer: %d\n", err);
            break;
        }
        gettimeofday(&end, NULL);
        if ((end.tv_sec - start.tv_sec) >= 5)
        {
            if (port_distribution)
                print_top_5_keys();
            gettimeofday(&start, NULL);
        }
    }

cleanup:
    /* Clean up */
    ring_buffer__free(udp_rb);
    ring_buffer__free(tcp_rb);
    ring_buffer__free(tcp_output_rb);
    ring_buffer__free(port_events_rb);
    ring_buffer__free(perf_map);
    ring_buffer__free(trace_all_drop);
    probe_bpf__destroy(skel);

    return err < 0 ? -err : 0;
}
