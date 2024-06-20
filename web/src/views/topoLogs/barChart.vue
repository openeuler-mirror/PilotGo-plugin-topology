<template>
  <div id="main" ref="chartDom"></div>
</template>

<script setup lang="ts">
import { markRaw, nextTick, onMounted, ref, watchEffect } from 'vue';
import bar_option from './chart_options';
import * as echarts from 'echarts';
import type { logData } from '@/types/index';

const chartDom = ref(null);
const myChart = ref<any>(null);
let isSecondClick = ref(false);

let props = defineProps({
  results: {
    type: Array as () => logData[],
    required: false,
    default: []
  }
})
let emit = defineEmits(["firstClick", "secondClick"]);

onMounted(() => {
  myChart.value = markRaw(echarts.init(chartDom.value!))
  // 柱状图点击事件
  myChart.value.on('click', function (bar_item_params: any) {
    // console.log(`图的信息：${bar_item_params},图的纵坐标：${bar_item_params.value}`);
    if (!isSecondClick.value) {
      // 第一次点击
      emit('firstClick', bar_item_params)
      isSecondClick.value = true;
    } else {
      // 第二次点击
      isSecondClick.value = true;
      emit('secondClick', bar_item_params)
    }
  });
  window.addEventListener('resize', resize)
})

watchEffect(() => {
  if (props.results) {
    nextTick(() => {
      handleBarData(props.results);
    })
  }
})

const resize = () => {
  nextTick(() => {
    myChart.value.resize()
  })
}

interface serieItem {
  name: string,
  type: string,
  stack: string,
  data: any
}
// 处理柱状图数据
const handleBarData = (_data: logData[]) => {
  if (!_data.length) return;
  bar_option.xAxis!.data = _data[0].data.map((item: any) => item[0]);
  let series: any = [];
  let seriesI: serieItem = {
    name: '',
    type: 'bar',
    stack: '',
    data: []
  };

  _data.forEach((item: any) => {
    seriesI.stack = 'A';
    seriesI.name = item.name;
    seriesI.data = item.data.map(function (d_item: any) {
      return d_item[1];
    });
    series.push(JSON.parse(JSON.stringify(seriesI)));
  })

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