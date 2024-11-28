/* 
 * Copyright (c) KylinSoft  Co., Ltd. 2024.All rights reserved.
 * PilotGo-plugin-topology licensed under the Mulan Permissive Software License, Version 2. 
 * See LICENSE file for more details.
 * Author: zhaozhenfang <zhaozhenfang@kylinos.cn>
 * Date: Wed Jul 17 16:38:05 2024 +0800
 */

export let startTime = (new Date() as any) / 1000 - 60 * 60 * 2;
export let endTime = (new Date() as any) / 1000;

// 过滤prometheus接口返回的数据
export const filterProm = (res: any) => {
  if (res && res.data.status === 'success' && res.data.data.result.length > 0) {
    return res.data.data.result.length === 1 ? res.data.data.result[0] : res.data.data.result;
  } else {
    return [];
  }
}
// 深拷贝数组
export const deepClone = (res: any) => {
  return JSON.parse(JSON.stringify(res))
}

// 测试数组是否为空,为空返回[],不为空返回有值元素的 ogject[]
export const filterNonEmptyObjects = (arr:any[]) => {  
  return arr.reduce((acc, innerArr) => {  
    // 遍历内部数组中的每个元素  
    if (!Array.isArray(innerArr)) {
      acc.push(innerArr);
    } else {
      innerArr.forEach((item:any) => {  
        // 检查当前项是否是一个非空对象  
        if (item && typeof item === 'object' && Object.keys(item).length > 0) {  
          // 如果是非空对象，则添加到累加器数组中  
          acc.push(item);  
        }  
      });  
    }
    return acc;  
  }, []); // 初始累加器数组为空数组  
} 

// 给对象数组的每个元素嵌套数组
export const nestedArray = (arr:any) => {
  return arr.map((item:any) => {
    if (Array.isArray(item)) {
      return item;
    } else {
      return [item];
    }
  })
}


// 处理字节类型数据
export const handle_byte = (value: string, float: number, unit: string) => {
  let value2unit: string = '0.00';
  switch (unit) {
    case 'GB':
      value2unit = (parseFloat(value) / 1024 / 1024 / 1024).toFixed(float);
      break;
    case 'KB':
      value2unit = (parseFloat(value) / 1024).toFixed(float);
      break;

    default:
      // MB
      value2unit = (parseFloat(value) / 1024 / 1024).toFixed(float);
      break;
  }
  return value2unit;
}

// echart颜色表
let e_colors = ['#2988f1', '#f8b238', '#b477f7', '#00c69e', '#ff79c2', '#40b85b', '#e24d42', '#ba43a9'];

// 柱状图
export const bar_opt = {
  color: e_colors,
  xAxis: {
    type: 'category',
    data: ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun']
  },
  yAxis: {
    type: 'value'
  },
  series: [
    {
      data: [120, 200, 150, 80, 70, 110, 130],
      type: 'bar',
      showBackground: true,
      backgroundStyle: {
        color: 'rgba(180, 180, 180, 0.2)'
      }
    }
  ]
}

// 折线图
export const line_opt = {
  color: e_colors,
  tooltip: {
    show: true,
    trigger: 'axis',
    confine: true,
    transitionDuration: 0.4,
    backgroundColor: 'rgba(255,255,255,1)',
    extraCssText: 'min-height:130px;',
    formatter: () => { },
    axisPointer: {
      animation: false
    }
  },
  legend: {
    show: true,
    orient: 'horizontal',
    right: 10,
    top: 0,
    textStyle: {
      color: '#888'
    },
    icon: 'circle',
  },
  dataZoom: [
    {
      show: true,
      type: 'slider',
      realtime: true,
      start: 65,
      end: 100,
      height: 18,
      showDetail: false,
      backgroundColor: "#fff",
      borderColor: "#DDE0E7",
      brushSelect: false,
      dataBackground: {
        lineStyle: {
          color: '#000'
        }
      },
      handleSize: "90%",
      // 长方形： path://M306.1,413c0,2.2-1.8,4-4,4h-59.8c-2.2,0-4-1.8-4-4V200.8c0-2.2,1.8-4,4-4h59.8c2.2,0,4,1.8,4,4V413z
      // 圆形带横条：M10.7,11.9v-1.3H9.3v1.3c-4.9,0.3-8.8,4.4-8.8,9.4c0,5,3.9,9.1,8.8,9.4v1.3h1.3v-1.3c4.9-0.3,8.8-4.4,8.8-9.4C19.5,16.3,15.6,12.2,10.7,11.9z M13.3,24.4H6.7V23h6.6V24.4z M13.3,19.6H6.7v-1.4h6.6V19.6z
      // 小圆形：path://M512,512m-448,0a448,448,0,1,0,896,0a448,448,0,1,0,-896,0Z
      handleIcon: "path://M306.1,413c0,2.2-1.8,4-4,4h-59.8c-2.2,0-4-1.8-4-4V200.8c0-2.2,1.8-4,4-4h59.8c2.2,0,4,1.8,4,4V413z", //icon图标
      handleStyle: {
        // color: '#B9C2CB',
      }
    }
  ],
  grid: {
    top: '12%',
    left: '2%',
    right: '2%',
    bottom: '18%',
    containLabel: true
  },
  xAxis: {
    type: 'time',
    splitLine: {
      show: false
    }
  },
  yAxis: {
    type: 'value',
    splitLine: {
      show: true,
      lineStyle: {
        color: ['#eee']
      }
    },
    min: 0,
    boundaryGap: [0, '100%'],
    axisLabel: {
      formatter: '{value}'
    }
  },
  series: [{
    name: '',
    type: 'line',
    smooth: false,
    showSymbol: false,
    z: 1,
    zlevel: 1,
    lineStyle: {
      width: 2
    },
    areaStyle: {
      opacity: 0.1,
    },
    data: [],
  }]
}
// 仪表盘
export const gauge_opt = {
  series: [
    {
      type: 'gauge',
      center: ['50%', '70%'],
      radius: '110%',
      startAngle: 180,
      endAngle: 0,
      min: 0,
      max: 100,
      tooltip: {

      },
      progress: {
        show: true,
        roundCap: true,
        width: 10,
        itemStyle: {
          color: {
            type: 'linear',
            x: 0,
            y: 0,
            x2: 0,
            y2: 1,
            colorStops: [{
              offset: 0, color: '#2988f1' // 0% 处的颜色#6bccff
            }, {
              offset: 1, color: '#6bccff' // 100% 处的颜色
            }],
            global: false // 缺省为 false
          }
        }
      },
      pointer: {
        show: false,
        itemStyle: {
          color: 'inherit'
        }
      },
      axisLine: {
        show: true,
        roundCap: true,
        lineStyle: {
          width: 10,
        }
      },
      axisTick: {
        show: false,
      },
      splitLine: {
        show: false
      },
      axisLabel: {
        show: false
      },
      anchor: {
        show: false
      },
      title: {
        show: false
      },
      detail: {
        valueAnimation: true,
        offsetCenter: [0, 0],
        fontSize: 20,
        fontWeight: 'bolder',
        formatter: '{value}',
        color: '#222'
      },
      data: [
        {
          value: 64.40
        }
      ]
    },
  ]
};