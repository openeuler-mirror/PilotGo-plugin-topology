import { defineStore } from 'pinia';
export let startTime = (new Date() as any) / 1000 - 60 * 60 * 2;
export let endTime = (new Date() as any) / 1000;
export const useLayoutStore = defineStore('layoutOption', {
  state: () => {
    return {
      layout_option: [
        {
          x: 0, y: 0, w: 1, h: 1, i: '0',
          static: false, display: false, title: 'CPU总使用率',
          query: {
            sqls: [{ sql: '100 - (avg(irate(node_cpu_seconds_total{instance="{macIp}",mode="idle"}[5m])) * 100)' }],
            type: 'gauge', range: false, isChart: true, interval: 5,
            target: 'percent_series', unit: '%', float: 2, min: 0, max: 100,
            color: [
              [0.5, '#67e0e3'],
              [0.8, '#E6A23C'],
              [1, '#fd666d']
            ]
          }
        },
        {
          x: 1, y: 0, w: 1, h: 1, i: '1',
          static: false, display: false, title: '内存使用率',
          query: {
            sqls: [{ sql: '(1 - (node_memory_MemAvailable_bytes{instance="{macIp}"} / (node_memory_MemTotal_bytes{instance="{macIp}"})))* 100' }],
            type: 'gauge', range: false, isChart: true, interval: 5,
            target: 'percent_series', unit: '%', float: 2, min: 0, max: 100,
            color: [
              [0.8, '#67e0e3'],
              [0.9, '#E6A23C'],
              [1, '#fd666d']
            ]
          }
        },
        {
          x: 0, y: 1, w: 3, h: 2, i: '2',
          static: false, display: false, title: '系统平均负载',
          query: {
            type: 'line', range: true, isChart: true, interval: 5,
            target: 'value_series', unit: '', float: 2, min: 0, max: null,
            sqls: [
              {
                sql: 'node_load1{instance="{macIp}"}',
                start: startTime,
                end: endTime,
                series_name: '1分钟'
              },
              {
                sql: 'node_load5{instance="{macIp}"}',
                start: startTime,
                end: endTime,
                series_name: '5分钟'
              },
              {
                sql: 'node_load15{instance="{macIp}"}',
                start: startTime,
                end: endTime,
                series_name: '15分钟'
              },
            ],
          }
        },
        {
          x: 0, y: 2, w: 3, h: 2, i: '3',
          static: false, display: false, title: '内存信息',
          query: {
            type: 'line', range: true, isChart: true,
            target: 'byte2GB_series', unit: 'GiB', float: 2, min: 0, max: null,
            sqls: [
              {
                sql: 'node_memory_MemTotal_bytes{instance="{macIp}"}',
                start: startTime,
                end: endTime,
                step: 15,
                series_name: '总内存',
              },
              {
                sql: 'node_memory_MemTotal_bytes{instance="{macIp}"} - node_memory_MemAvailable_bytes{instance="{macIp}"}',
                start: startTime,
                end: endTime,
                step: 15,
                series_name: '已用',
              },
              {
                sql: 'node_memory_MemAvailable_bytes{instance="{macIp}"}',
                start: startTime,
                end: endTime,
                step: 15,
                series_name: '可用',
              },

            ]
          }
        },
      ],
      process_layout_option: [
        {
          x: 0, y: 0, w: 3, h: 2, i: '0',
          static: false, display: false, title: '***chart',
          query: {
            type: 'line', range: true, isChart: true,
            target: 'byte2GB_series', unit: 'GiB', float: 2, min: 0, max: null,
            sqls: [
              {
                sql: 'node_memory_MemTotal_bytes{instance="{macIp}"}',
                start: startTime,
                end: endTime,
                step: 15,
                series_name: '总内存',
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
