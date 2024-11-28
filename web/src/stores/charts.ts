/* 
 * Copyright (c) KylinSoft  Co., Ltd. 2024.All rights reserved.
 * PilotGo-plugin-topology licensed under the Mulan Permissive Software License, Version 2. 
 * See LICENSE file for more details.
 * Author: zhao_zhen_fang <zhaozhenfang@kylinos.cn>
 * Date: Mon Nov 13 18:00:47 2023 +0800
 */
import { defineStore } from 'pinia';
export let startTime = (new Date() as any) / 1000 - 60 * 60 * 2;
export let endTime = (new Date() as any) / 1000;
export const useLayoutStore = defineStore('layoutOption', {
  state: () => {
    return {
      layout_option: [
        {
          x: 0, y: 0, w: 2, h: 1, i: '0',
          static: true, display: true, title: '运行时间',
          query: {
            sqls: [{ sql: '(time()-node_boot_time_seconds{group="",instance="{macIp}"})/(60*60*24)' }],
            type: 'value', range: false, isValue: true, interval: 5,
            target: 'value_series', unit: '天', float: 2
          }
        },
        {
          x: 2, y: 0, w: 2, h: 1, i: '1',
          static: true, display: true, title: 'CPU核数',
          query: {
            sqls: [{ sql: 'count(count(node_cpu_seconds_total{group="",instance="{macIp}",mode="system"}) by (cpu))' }],
            type: 'value', range: false, isValue: true, interval: 5,
            target: 'value_series', unit: '个', float: 0
          }
        },
        {
          x: 4, y: 0, w: 2, h: 1, i: '2',
          static: true, display: true, title: '内存总量',
          query: {
            sqls: [{ sql: 'node_memory_MemTotal_bytes{group="",instance="{macIp}"}' }],
            type: 'value', range: false, isValue: true, interval: 5,
            target: 'byte2GB_series', unit: 'GiB', float: 2,

          }
        },
        /* {
          x: 6, y: 0, w: 2, h: 1, i: '3',
          static: true, display: true, title: '当前打开的文件描述符',
          query: {
            sqls: [{ sql: 'node_filefd_allocated{group="",instance="{macIp}"}' }],
            type: 'value', range: false, isValue: true, interval: 5,
            target: 'num_series', unit: 'K', float: 2, min: 0, max: 9,
            color: []
          }
        }, */
        {
          x: 0, y: 1, w: 2, h: 2, i: '4',
          static: true, display: true, title: 'CPU总使用率',
          query: {
            sqls: [{ sql: '100 - (avg(irate(node_cpu_seconds_total{group="",instance="{macIp}",mode="idle"}[5m])) * 100)' }],
            type: 'gauge', range: false, isChart: true, interval: 5,
            target: 'percent_series', unit: '%', float: 2, min: 0, max: 100,
            color: [{
              offset: 0, color: '#2988f1'
            }, {
              offset: 1, color: '#6bccff'
            }]
          }
        },
        {
          x: 2, y: 1, w: 2, h: 2, i: '5',
          static: true, display: true, title: 'CPU iowait',
          query: {
            sqls: [{ sql: 'avg(irate(node_cpu_seconds_total{group="",instance="{macIp}",mode="iowait"}[5m])) * 100' }],
            type: 'gauge', range: false, isChart: true, interval: 5,
            target: 'percent_series', unit: '%', float: 2, min: 0, max: 100,
            color: [{
              offset: 0, color: '#965af2'
            }, {
              offset: 1, color: '#cc8dfa'
            }]
          }
        },
        {
          x: 4, y: 1, w:2, h: 2, i: '6',
          static: true, display: true, title: '内存使用率',
          query: {
            sqls: [{ sql: '(1 - (node_memory_MemAvailable_bytes{group="",instance="{macIp}"} / (node_memory_MemTotal_bytes{group="",instance="{macIp}"})))* 100' }],
            type: 'gauge', range: false, isChart: true, interval: 5,
            target: 'percent_series', unit: '%', float: 2, min: 0, max: 100,
            color: [{
              offset: 0, color: '#eb9509'
            }, {
              offset: 1, color: '#febf50'
            }]
          }
        },
        /* {
          x: 6, y: 1, w: 2, h: 2, i: '7',
          static: true, display: true, title: '根分区使用率',
          query: {
            sqls: [{ sql: '100 - ((node_filesystem_avail_bytes{group="",instance="{macIp}",mountpoint="/",fstype=~"ext4|xfs"} * 100) / node_filesystem_size_bytes {group="",instance="{macIp}",mountpoint="/",fstype=~"ext4|xfs"})' }],
            type: 'gauge', range: false, isChart: true, interval: 5,
            target: 'percent_series', unit: '%', float: 2, min: 0, max: 100,
            color: [{
              offset: 0, color: '#00c59d'
            }, {
              offset: 1, color: '#1de3cf'
            }]
          }
        }, */
        {
          x: 0, y: 3, w: 8, h: 3, i: '8',
          static: false, display: true, title: '系统平均负载',
          query: {
            type: 'line', range: true, isChart: true, interval: 5,
            target: 'value_series', unit: '', float: 2, min: 0, max: null,
            sqls: [
              {
                sql: 'node_load1{group="",instance="{macIp}"}',
                start: startTime,
                end: endTime,
                series_name: '1分钟'
              },
              {
                sql: 'node_load5{group="",instance="{macIp}"}',
                start: startTime,
                end: endTime,
                series_name: '5分钟'
              },
              {
                sql: 'node_load15{group="",instance="{macIp}"}',
                start: startTime,
                end: endTime,
                series_name: '15分钟'
              },
            ],
          }
        },
        {
          x: 0, y: 3, w: 8, h: 3, i: '9',
          static: false, display: true, title: '内存信息',
          query: {
            type: 'line', range: true, isChart: true,
            target: 'byte2GB_series', unit: 'GiB', float: 2, min: 0, max: null,
            sqls: [
              {
                sql: 'node_memory_MemTotal_bytes{group="",instance="{macIp}"}',
                start: startTime,
                end: endTime,
                step: 15,
                series_name: '总内存',
              },
              {
                sql: 'node_memory_MemTotal_bytes{group="",instance="{macIp}"} - node_memory_MemAvailable_bytes{group="",instance="{macIp}"}',
                start: startTime,
                end: endTime,
                step: 15,
                series_name: '已用',
              },
              {
                sql: 'node_memory_MemAvailable_bytes{group="",instance="{macIp}"}',
                start: startTime,
                end: endTime,
                step: 15,
                series_name: '可用',
              },

            ]
          }
        },
        {
          x: 0, y: 6, w: 8, h: 3, i: '10',
          static: false, display: true, title: 'cpu使用率',
          query: {
            type: 'line', range: true, isChart: true,
            target: 'percent_series', unit: '%', float: 2, min: 0, max: null,
            sqls: [
              {
                sql: 'avg(irate(node_cpu_seconds_total{group="",instance="{macIp}",mode="system"}[1m])) by (instance)',
                start: startTime,
                end: endTime,
                step: 15,
                series_name: '系统cpu使用率',
              },
              {
                sql: 'avg(irate(node_cpu_seconds_total{group="",instance="{macIp}",mode="user"}[1m])) by (instance)',
                start: startTime,
                end: endTime,
                step: 15,
                series_name: '用户cpu使用率',
              },
              {
                sql: 'avg(irate(node_cpu_seconds_total{group="",instance="{macIp}",mode="idle"}[1m])) by (instance)',
                start: startTime,
                end: endTime,
                step: 15,
                series_name: '单核cpu空闲率',
              },
              {
                sql: 'avg(irate(node_cpu_seconds_total{group="",instance="{macIp}",mode="iowait"}[1m])) by (instance)',
                start: startTime,
                end: endTime,
                step: 15,
                series_name: '磁盘io使用率',
              },
              /* {
                sql: 'irate(node_disk_io_time_seconds_total{group="",instance="{macIp}",}[1m])',
                start: startTime,
                end: endTime,
                step: 15,
                series_name: '',
              }, */
            ]
          }
        },
        {
          x: 0, y: 9, w: 8, h: 3, i: '11',
          static: false, display: true, title: '磁盘总空间',
          query: {
            type: 'table', range: false, isChart: false, interval: 5,
            target: 'more_query',
            sqls: [{
              sql: 'node_filesystem_size_bytes{group="",instance="{macIp}",fstype=~"ext4|xfs"}/10^9',
              columnList: [
                { prop: "filesystem", label: '文件系统' },
                { prop: "zone", label: '分区' },
                { prop: "size", label: '空间大小(GB)' },
              ],
              columnValue: [
                { prop: 'filesystem', value: 'item.metric.fstype', type: '' },
                { prop: 'zone', value: 'item.metric.mountpoint', type: '' },
                { prop: 'size', value: 'item.value[1]', type: 'float' }
              ]
            }]
          }
        },
        {
          x: 0, y: 9, w: 8, h: 3, i: '12',
          static: false, display: true, title: '各分区可用空间',
          query: {
            type: 'table', range: false, isChart: false, interval: 5,
            target: 'more_query',
            sqls: [{
              sql: 'node_filesystem_avail_bytes {group="",instance="{macIp}",fstype=~"ext4|xfs"}',
              columnList: [
                { prop: "filesystem", label: '文件系统' },
                { prop: "zone", label: '分区' },
                { prop: "avail", label: '可用空间(GB)' },
              ],
              columnValue: [
                { prop: 'filesystem', value: 'item.metric.fstype', type: '' },
                { prop: 'zone', value: 'item.metric.mountpoint', type: '' },
                { prop: 'avail', value: 'item.value[1]', type: 'byte', }
              ]
            },
            {
              sql: '1-(node_filesystem_free_bytes{group="",instance="{macIp}",fstype=~"ext4|xfs"} / node_filesystem_size_bytes{group="",instance="{macIp}",fstype=~"ext4|xfs"})',
              columnList: [
                { prop: "use", label: '使用率' },
              ],
              columnValue: [
                { prop: 'use', value: 'item.value[1]', type: 'percent', }
              ]
            }]
          }
        },
        {
          x: 0, y: 12, w: 8, h: 3, i: '13',
          static: false, display: true, title: '磁盘读取容量',
          query: {
            type: 'line', range: true, isChart: true,
            target: 'byte2KB_series', unit: 'kB/s', float: 2, min: 0, max: null,
            sqls: [
              {
                sql: 'rate(node_disk_read_bytes_total{group="",instance="{macIp}"}[1m])',
                start: startTime,
                end: endTime,
                step: 15,
                series_name: '读取',
              },


            ]
          }
        },
        {
          x: 0, y: 12, w: 8, h: 3, i: '14',
          static: false, display: true, title: '磁盘写入容量',
          query: {
            type: 'line', range: true, isChart: true,
            target: 'byte2KB_series', unit: 'kB/s', float: 2, min: 0, max: null,
            sqls: [
              {
                sql: 'rate(node_disk_written_bytes_total{group="",instance="{macIp}"}[1m])',
                start: startTime,
                end: endTime,
                step: 15,
                series_name: '写入',
              },

            ]
          }
        },
        {
          x: 0, y: 15, w: 8, h: 3, i: '15',
          static: false, display: true, title: 'TCP连接情况',
          query: {
            type: 'line', range: true, isChart: true,
            target: '', unit: '', float: 2, min: 0, max: null,
            sqls: [
              {
                sql: 'node_netstat_Tcp_CurrEstab{group="",instance="{macIp}"}',
                start: startTime,
                end: endTime,
                step: 15,
                series_name: 'ESTABLISHED',
              },
              {
                sql: 'node_sockstat_TCP_tw{group="",instance="{macIp}"}',
                start: startTime,
                end: endTime,
                step: 15,
                series_name: 'TCP_tw',
              },
              {
                sql: 'irate(node_netstat_Tcp_ActiveOpens{group="",instance="{macIp}"}[1m])',
                start: startTime,
                end: endTime,
                step: 15,
                series_name: 'ActiveOpens',
              },
              {
                sql: 'irate(node_netstat_Tcp_PassiveOpens{group="",instance="{macIp}"}[1m])',
                start: startTime,
                end: endTime,
                step: 15,
                series_name: 'PassiveOpens',
              },
              {
                sql: 'node_sockstat_TCP_alloc{group="",instance="{macIp}"}',
                start: startTime,
                end: endTime,
                step: 15,
                series_name: 'TCP_alloc',
              },
              {
                sql: 'node_sockstat_TCP_inuse{group="",instance="{macIp}"}',
                start: startTime,
                end: endTime,
                step: 15,
                series_name: 'TCP_inuse',
              },
            ]
          }
        },
      ],
    };
  },
  /* persist: {
    enabled: true, // 开启存储
    strategies: [
      { storage: localStorage, paths: ["layout_option"] },
    ]
  }, */
  getters: {},
  actions: {
    initLayout(layout: any) {
      this.layout_option = layout;
    },
    addLayout(layout: any) {
      this.layout_option.push(layout);
    }
  }
});
