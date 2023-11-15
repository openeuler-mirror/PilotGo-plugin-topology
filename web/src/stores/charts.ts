import { defineStore } from 'pinia';
export let startTime = (new Date() as any) / 1000 - 60 * 60 * 2;
export let endTime = (new Date() as any) / 1000;
export const useLayoutStore = defineStore('layoutOption', {
  state: () => {
    return {
      layout_option: [
        {
          x: 0, y: 0, w: 2, h: 4, i: '3',
          static: true, display: true, title: 'CPU总使用率',
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
        // {
        //   x: 1, y: 0, w: 2, h: 4, i: '5',
        //   static: false, display: true, title: '内存使用率',
        //   query: {
        //     sqls: [{ sql: '(1 - (node_memory_MemAvailable_bytes{instance="{macIp}"} / (node_memory_MemTotal_bytes{instance="{macIp}"})))* 100' }],
        //     type: 'gauge', range: false, isChart: true, interval: 5,
        //     target: 'percent_series', unit: '%', float: 2, min: 0, max: 100,
        //     color: [
        //       [0.8, '#67e0e3'],
        //       [0.9, '#E6A23C'],
        //       [1, '#fd666d']
        //     ]
        //   }
        // },
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
