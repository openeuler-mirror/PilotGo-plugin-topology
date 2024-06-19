<template>
  <div id="main" ref="chartDom"></div>
</template>

<script setup lang="ts">
import { markRaw, nextTick, onMounted, ref } from 'vue';
import bar_option from './chart_options';
import * as echarts from 'echarts';

const chartDom = ref(null);
const myChart = ref<any>(null);

let props = defineProps({
  logs: {
    type: Array,
    required: false,
    default: []
  }
})

onMounted(() => {
  myChart.value = markRaw(echarts.init(chartDom.value!))
  /* option.value = bar_option;
  myChart.value.setOption(option.value, true) */
  handleBarData();
  window.addEventListener('resize', resize)
})
const resize = () => {
  nextTick(() => {
    myChart.value.resize()
  })
}

declare type serieItem = {
  name: string,
  type: string,
  stack: string,
  data: any
}
// 处理柱状图数据
const handleBarData = () => {
  let series: any = [];
  let seriesI: serieItem = {
    name: '',
    type: 'bar',
    stack: '',
    data: []
  };
  bar_option.xAxis!.data = props.logs.map((item: any) => item[0]);
  // 对接口数据进行处理
  seriesI.name = '类别1';
  seriesI.stack = 'A';
  seriesI.data = props.logs.map(function (item: any) {
    return item[1];
  });
  series.push(seriesI);
  setBorderRadius(series);
  bar_option.series = series;
  myChart.value.setOption(bar_option, true)
}

// 设置柱状图圆角
const setBorderRadius = (series: any) => {
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


</script>

<style scoped>
#main {
  width: 1300px;
  height: 300px;
  margin: 0 auto;
}
</style>