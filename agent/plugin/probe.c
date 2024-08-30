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
#include <unistd.h>
#include "probe.h"
#include "probe.skel.h"
#include "probe.h"

static volatile bool exiting = false;
int udp_info = 0, tcp_status_info = 0, tcp_output_info = 0, protocol_info = 0;
struct protocol_stats proto_stats[256] = {0};
time_t start_time;
int interval = 5; // 每5 秒计算一次

const char argp_program_doc[] = "Trace time delay in network subsystem \n";

static const struct argp_option opts[] = {
    {"udp", 'u', 0, 0, "trace the udp message"},
    {"tcp_status_info", 't', 0, 0, "trace the tcp states"},
    {"tcp_output_info", 'o', 0, 0, "trace the tcp flow"},
    {"protocol_info", 'p', 0, 0, "trace the tcp flow"},
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
        if (proto_stats[i].rx_count >= last_rx[i]) {
            delta_rx[i] = proto_stats[i].rx_count - last_rx[i];
        } else {
            delta_rx[i] = proto_stats[i].rx_count;  
        }

        if (proto_stats[i].tx_count >= last_tx[i]) {
            delta_tx[i] = proto_stats[i].tx_count - last_tx[i];
        } else {
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

static int print_count_protocol_use(void *ctx, void *packet_info, size_t data_sz)
{
    const struct packet_info *pack_protocol_info = (const struct packet_info *)packet_info;
    if (protocol_info)
    {
        proto_stats[pack_protocol_info->proto].rx_count += pack_protocol_info->count.rx_count;
        proto_stats[pack_protocol_info->proto].tx_count += pack_protocol_info->count.tx_count;
    }
    return 0;
}

int main(int argc, char **argv)
{
    struct probe_bpf *skel;
    int err = 0;
    struct ring_buffer *udp_rb = NULL;
    struct ring_buffer *tcp_rb = NULL;
    struct ring_buffer *tcp_output_rb = NULL;
    struct ring_buffer *port_events = NULL;
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
    port_events = ring_buffer__new(bpf_map__fd(skel->maps.port_events), print_count_protocol_use, NULL, NULL);
    if (!port_events)
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
    if (protocol_info)
    {
        printf("==========Proportion of each agreement==========\n");
    }
    start_time = time(NULL);
    while (!exiting)
    {
        err = ring_buffer__poll(udp_rb, 100 /* timeout, ms */);
        err = ring_buffer__poll(tcp_rb, 100 /* timeout, ms */);
        err = ring_buffer__poll(tcp_output_rb, 100 /* timeout, ms */);
        err = ring_buffer__poll(port_events, 100 /* timeout, ms */);

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
    }

cleanup:
    /* Clean up */
    ring_buffer__free(udp_rb);
    ring_buffer__free(tcp_rb);
    ring_buffer__free(tcp_output_rb);
    ring_buffer__free(port_events);
    probe_bpf__destroy(skel);

    return err < 0 ? -err : 0;
}
