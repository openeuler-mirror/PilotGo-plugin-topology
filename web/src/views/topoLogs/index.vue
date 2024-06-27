<!-- elk插件展示日志页面 -->
<template>
  <div class="log_containt shadow">
    <el-tabs v-model="activeName" class="log_containt_tabs " @tab-click="handleClick">
      <el-tab-pane label="日志" name="log" :lazy="true">
        <div class="log_containt_query">
          <!-- 1.简单语句拼写 -->
          <div class="category">
            分类类别：
          </div>
          <!-- 2.时间范围选择 -->
          <div class="time">
            时间范围：
            <el-date-picker v-model="log_time" type="datetimerange" range-separator="To" start-placeholder="Start date"
              end-placeholder="End date" @change="ChangeTimeRange" @clear="ChangeTimeRange" />
          </div>
          <!-- 3.按钮 -->
          <el-button type="primary" :icon="Search" @click="handleGetData()">搜索</el-button>
        </div>
        <barChart :results="logs" @firstClick="getLogData" @secondClick="handleShowLog" />
      </el-tab-pane>
      <el-tab-pane label="事件" name="event" :lazy="true">
        <div class="log_containt_query">
          <!-- 1.简单语句拼写 -->
          <div class="category">
            分类类别：
          </div>
          <!-- 2.时间范围选择 -->
          <div class="time">
            时间范围：
            <el-date-picker v-model="event_time" type="datetimerange" range-separator="To"
              start-placeholder="Start date" end-placeholder="End date" @change="ChangeEventTimeRange"
              @clear="ChangeEventTimeRange" />
          </div>
          <!-- 3.按钮 -->
          <el-button type="primary" :icon="Search" @click="handleGetData()">搜索</el-button>
        </div>
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
import { Search } from '@element-plus/icons-vue'

import barChart from './barChart.vue';
import logStream from './logStream.vue'
import { useTopoStore } from '@/stores/topo';
import type { logData } from '@/types/index';
import { getELKLogData } from '@/request/elk';
import { calculate_interval } from './utils';
// import result, { host_result, log_query } from './test';

const activeName = ref('log');
const logs = ref([] as logData[]);
const event_logs = ref([] as logData[]);

const dialog = ref(false);
const title = ref('日志详情');
const log_type = ref('log');

onMounted(() => {
  getLogData();
})

const log_time = ref<[Date, Date]>([
  new Date(new Date().getTime() - 2 * 60 * 60 * 1000),
  new Date(),
])
const event_time = ref<[Date, Date]>([
  new Date(new Date().getTime() - 2 * 60 * 60 * 1000),
  new Date(),
])
// 时间选择框触发选择
const ChangeTimeRange = (value: any) => {
  if (value) {
    log_time.value[0] = value[0]
    log_time.value[1] = value[1]
  }
}
const ChangeEventTimeRange = (value: any) => {
  if (value) {
    event_time.value[0] = value[0]
    event_time.value[1] = value[1]
  }
}

// 处理请求体参数
const handleParams = (_id?: string) => {
  let log_query = {
    id: 'log_timeaxis',
    params: {
      query_data_stream_dataset: "system.syslog",
      query_range_gte: 1719226716185,
      query_range_lte: 1719226836185,
      aggs_field: "host.hostname",
      size: 0,
      fixed_interval: "5s"
    }
  }
  // calculate_interval(log_time.value[0], log_time.value[1])


  switch (log_type.value) {
    case 'log':
      log_query.params.query_range_gte = log_time.value[0].getTime();
      log_query.params.query_range_lte = log_time.value[1].getTime();
      log_query.params.fixed_interval = calculate_interval(log_time.value[0], log_time.value[1]) + 's';
      break;

    default:
      log_query.params.query_range_gte = event_time.value[0].getTime();
      log_query.params.query_range_lte = event_time.value[1].getTime();
      log_query.params.fixed_interval = calculate_interval(event_time.value[0], event_time.value[1]) + 's';
      break;
  }
  return log_query;

}

// 请求日志数据
const getLogData = () => {
  // logs.value = result;
  switch (log_type.value) {
    case 'log':
      getELKLogData(handleParams()).then(res => {
        if (res.data.code === 200) {
          logs.value = res.data.data;
        }
      })
      break;

    default:

      break;
  }

}

// 切换tab事件
const handleClick = (_tab: TabsPaneContext, _event: Event) => {
  log_type.value = _tab.props.name as string;
}

// 点击搜索按钮
const handleGetData = () => {
  switch (log_type.value) {
    case 'log':
      getLogData();
      break;

    default:
      // getEventData(type);
      break;
  }
}

// 相应点击进程柱状图事件
const handleShowLog = (_process_info: any) => {
  console.log('进程信息：', _process_info)
  if (_process_info) {
    dialog.value = true;
    title.value = _process_info.seriesName + '日志流';
  }
}

watch(() => useTopoStore().node_log_id, (new_node_id) => {
  if (new_node_id) {
    // 发送接口请求
    getLogData();
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
    width: 98%;
    margin: 0 auto;
    height: 66px;
    display: flex;
    justify-content: space-between;
    align-items: center;

    .category {
      width: 22%;
    }

    .time {
      display: flex;
      align-items: center;
      width: 500px;
    }
  }

  &_tabs {
    width: 100%;
    height: 100%;
  }

}
</style>