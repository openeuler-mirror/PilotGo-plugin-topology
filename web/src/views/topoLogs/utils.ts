/* 
 * Copyright (c) KylinSoft  Co., Ltd. 2024.All rights reserved.
 * PilotGo-plugin-topology licensed under the Mulan Permissive Software License, Version 2. 
 * See LICENSE file for more details.
 * Author: zhaozhenfang <zhaozhenfang@kylinos.cn>
 * Date: Thu Jun 27 10:44:47 2024 +0800
 */
// 动态计算时间间隔
export const calculate_interval = (start: Date, end: Date) => {
  let time_interval: number;
  if (!start || !end) return;
  let time_range: number;
  time_range = (end.getTime() - start.getTime()) / 1000;
  time_interval = time_range / 100 <= 5 ? 5 : Math.ceil(time_range / 100);
  return time_interval;
}


// 设置柱状图圆角
export const setBorderRadius = (series: any) => {
  const stackInfo: any = {};
  for (let i = 0; i < series[0].data.length; ++i) {
    for (let j = 0; j < series.length; ++j) {
      const stackName = series[j].stack;
      if (!stackName) {
        continue;
      }
      if (!stackInfo[stackName]) {
        stackInfo[stackName] = {
          stackStart: [],
          stackEnd: []
        };
      }
      const info = stackInfo[stackName];
      const data = series[j].data[i];
      if (data && data !== '-') {
        if (info.stackStart[i] == null) {
          info.stackStart[i] = j;
        }
        info.stackEnd[i] = j;
      }
    }
  }
  for (let i = 0; i < series.length; ++i) {
    const data: any = series[i].data;
    const info = stackInfo[series[i].stack];
    for (let j = 0; j < series[i].data.length; ++j) {
      const isEnd = info.stackEnd[j] === i;
      const topBorder = isEnd ? 20 : 0;
      const bottomBorder = 0;
      data[j] = {
        value: data[j],
        itemStyle: {
          borderRadius: [topBorder, topBorder, bottomBorder, bottomBorder]
        }
      };
    }
  }
}

// 日志等级
export let levels = [
  {
    value: '0',
    label: 'emergency',
  },
  {
    value: '1',
    label: 'alert',
  },
  {
    value: '2',
    label: 'critical',
  },
  {
    value: '3',
    label: 'error',
  },
  {
    value: '4',
    label: 'warning',
  },
  {
    value: '5',
    label: 'notice',
  },
  {
    value: '6',
    label: 'informational',
  },
  {
    value: '7',
    label: 'debug',
  },
]