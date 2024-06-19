<!-- elk插件展示日志页面 -->
<template>
  <div class="log_containt">
    <el-tabs v-model="activeName" class="log_containt_tabs shadow" @tab-click="handleClick">
      <el-tab-pane label="日志" name="log">
        <barChart :results="logs" @firstClick="getLogData" @secondClick="handleShowLog" />
      </el-tab-pane>
      <el-tab-pane label="事件" name="event">
        <barChart :results="event_logs" />
      </el-tab-pane>
    </el-tabs>

  </div>

</template>

<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import type { TabsPaneContext } from 'element-plus'

import barChart from './barChart.vue';
import { useTopoStore } from '@/stores/topo';
import type { logData } from '@/types/index';

// import result, { host_result } from './test';

const activeName = ref('log');
const logs = ref([] as logData[]);
const event_logs = ref([] as logData[]);

onMounted(() => {
  getLogData();
})

// 请求日志数据
const getLogData = (_params?: any) => {
  if (_params) {
    // 点击主机bar
    // logs.value = host_result;
    return;
  }
  // logs.value = result;
}

const handleClick = (_tab: TabsPaneContext, _event: Event) => {
  // console.log(tab.props)

}

// 相应点击进程柱状图事件
const handleShowLog = (_process_info: any) => {
  // console.log('进程信息：', process_info)
}

watch(() => useTopoStore().node_log_id, (new_node_id) => {
  if (new_node_id) {
    // 发送接口请求
    getLogData({ log_id: new_node_id });
  }
})
</script>


<style scoped lang="scss">
.log_containt {
  width: 96%;
  height: 400px;
  position: fixed;
  bottom: 0;
  background-color: #fff;

  &_tabs {
    width: 100%;
    height: 100%;
  }

}
</style>