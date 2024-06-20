<!-- elk插件展示日志页面 -->
<template>
  <div class="log_containt shadow">
    <el-tabs v-model="activeName" class="log_containt_tabs " @tab-click="handleClick">
      <el-tab-pane label="日志" name="log">
        <barChart :results="logs" @firstClick="getLogData" @secondClick="handleShowLog" />
      </el-tab-pane>
      <el-tab-pane label="事件" name="event">
        <barChart :results="event_logs" />
      </el-tab-pane>
    </el-tabs>
    <el-dialog v-model="dialog" :title="title" width="80%">
      <logStream />
    </el-dialog>
  </div>

</template>

<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import type { TabsPaneContext } from 'element-plus'

import barChart from './barChart.vue';
import logStream from './logStream.vue'
import { useTopoStore } from '@/stores/topo';
import type { logData } from '@/types/index';

// import result, { host_result } from './test';

const activeName = ref('log');
const logs = ref([] as logData[]);
const event_logs = ref([] as logData[]);

const dialog = ref(false);
const title = ref('日志详情');

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
  if (_process_info) {
    dialog.value = true;
    title.value = _process_info.seriesName + '日志流';
  }
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
  height: 500px;
  position: fixed;
  bottom: 0;
  background-color: #fff;

  &_query {
    width: 100%;
    height: 10%;
  }

  &_tabs {
    width: 100%;
    height: 90%;
  }

}
</style>