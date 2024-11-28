<!--
 * Copyright (c) KylinSoft  Co., Ltd. 2024.All rights reserved.
 * PilotGo-plugin-topology licensed under the Mulan Permissive Software License, Version 2. 
 * See LICENSE file for more details.
 * Author: zhaozhenfang <zhaozhenfang@kylinos.cn>
 * Date: Tue Jun 18 10:33:42 2024 +0800
-->
<template>
  <div id="main" ref="chartDom"></div>
</template>

<script setup lang="ts">
import { markRaw, nextTick, onMounted, ref, watchEffect, watch } from 'vue';
import bar_option from './chart_options';
import * as echarts from 'echarts';
import type { logData } from '@/types/index';
import { setBorderRadius } from './utils'

const chartDom = ref(null);
const myChart = ref<any>(null);
let isSecondClick = ref(false);

let props = defineProps({
  results: {
    type: Array as () => logData[],
    required: false,
    default: []
  },
  clickChange: {
    type: String,
    required: false,
    default: 'first'
  }
})

let emit = defineEmits(["firstClick", "secondClick"]);

onMounted(() => {
  myChart.value = markRaw(echarts.init(chartDom.value!))
  // 柱状图点击事件
  myChart.value.on('click', function (bar_item_params: any) {
    // console.log(`图的信息：${bar_item_params},图的纵坐标：${bar_item_params.value}`);
    bar_option.title.text = bar_item_params.seriesName;
    if (!isSecondClick.value) {
      // 第一次点击
      emit('firstClick', bar_item_params)
      isSecondClick.value = true;
    } else if (props.results.length > 0) {
      // 第二次点击
      isSecondClick.value = true;
      emit('secondClick', bar_item_params)
    } else {
      console.log('error')
    }
  });
  window.addEventListener('resize', resize)
})
const resize = () => {
  nextTick(() => {
    myChart.value.resize()
  })
}

watchEffect(() => {
  if (props.clickChange === 'first') {
    isSecondClick.value = false;
    bar_option.title.text = '集群';
  }
  if (props.results) {
    nextTick(() => {
      handleBarData(props.results);
    })
  }
})


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

  _data.forEach((item: any) => {
    let seriesI: serieItem = {
      name: '',
      type: 'bar',
      stack: '',
      data: []
    };
    seriesI.stack = 'A';
    seriesI.name = item.name;
    seriesI.data = item.data;
    series.push(JSON.parse(JSON.stringify(seriesI)));
  })

  setBorderRadius(series);
  bar_option.series = series;
  myChart.value.setOption(bar_option, true)
}

</script>

<style scoped>
#main {
  width: 1300px;
  height: 350px;
  margin: 0 auto;
}
</style>