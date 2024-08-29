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
int udp_info = 0, tcp_status_info = 0;

const char argp_program_doc[] = "Trace time delay in network subsystem \n";

static const struct argp_option opts[] = {
    {"udp", 'u', 0, 0, "trace the udp message"},
    {"tcp_status_info", 't', 0, 0, "trace the tcp states"},
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

static void print_udp_event_info(const struct event *pack_info)
{
    char s_str[INET_ADDRSTRLEN];
    char d_str[INET_ADDRSTRLEN];

    if (pack_info->client_port == 0 || pack_info->server_port == 0 || pack_info->pid == 0)
    {
        return;
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
}

static void print_tcp_state_info(const struct event *pack_info)
{
    char s_str[INET_ADDRSTRLEN];
    char d_str[INET_ADDRSTRLEN];

    format_ip_address(pack_info->client_ip, s_str, sizeof(s_str));
    format_ip_address(pack_info->server_ip, d_str, sizeof(d_str));

    printf("%-20d %-20s %-20s %-20d %-20d %-20s %-20s %-20d\n",
           pack_info->pid,
           s_str,
           d_str,
           pack_info->client_port,
           pack_info->server_port,
           tcp_states[pack_info->oldstate],
           tcp_states[pack_info->newstate],
           pack_info->tran_time);
}

static int handle_event(void *ctx, void *packet_info, size_t data_sz)
{
    const struct event *pack_info = (const struct event *)packet_info;

    if (udp_info)
    {
        print_udp_event_info(pack_info);
    }

    if (tcp_status_info)
    {
        print_tcp_state_info(pack_info);
    }

    return 0;
}

int main(int argc, char **argv)
{
    struct probe_bpf *skel;
    int err = 0;
    struct ring_buffer *udp_rb = NULL;
    struct ring_buffer *tcp_rb = NULL;

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
    udp_rb = ring_buffer__new(bpf_map__fd(skel->maps.udp_rb), handle_event, NULL, NULL);
    if (!udp_rb)
    {
        err = -1;
        fprintf(stderr, "Failed to create ring buffer\n");
        goto cleanup;
    }
    tcp_rb = ring_buffer__new(bpf_map__fd(skel->maps.tcp_rb), handle_event, NULL, NULL);
    if (!tcp_rb)
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
        printf("%-20s %-20s %-20s %-20s %-20s %-20s %-20s %-20s \n", "Pid","Client_ip", "Server_ip", "Client_port", "Server_port", "oldstate", "newstate", "time/μs");
    }

    while (!exiting)
    {
        err = ring_buffer__poll(udp_rb, 100 /* timeout, ms */);
        err = ring_buffer__poll(tcp_rb, 100 /* timeout, ms */);
        /* Ctrl-C will cause -EINTR */
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
    probe_bpf__destroy(skel);

    return err < 0 ? -err : 0;
}
