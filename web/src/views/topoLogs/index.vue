<!--
 * Copyright (c) KylinSoft  Co., Ltd. 2024.All rights reserved.
 * PilotGo-plugin-topology licensed under the Mulan Permissive Software License, Version 2. 
 * See LICENSE file for more details.
 * Author: zhaozhenfang <zhaozhenfang@kylinos.cn>
 * Date: Wed Jun 12 17:21:06 2024 +0800
-->
<!-- elk插件展示日志页面 -->
<template>
  <div class="log_containt shadow">
    <el-tabs v-model="activeName" class="log_containt_tabs " @tab-click="handleClick">
      <el-tab-pane label="日志" name="log" :lazy="true">
        <div class="log_containt_query">
          <!-- 1.简单语句拼写 -->
          <!-- <div class="category">
            分类类别：
          </div> -->
          <!-- 2.时间范围选择 -->
          <div class="time">
            时间范围：
            <el-date-picker v-model="log_time" type="datetimerange" range-separator="To" start-placeholder="Start date"
              end-placeholder="End date" @change="ChangeTimeRange" @clear="ChangeTimeRange" />
          </div>
          <!-- 3.按钮 -->
          <div class="bt">
            <el-button type="primary" :icon="Search" @click="handleGetData()">搜索</el-button>
            <el-button type="primary" v-show="back_btn" :icon="Back" @click="handleReset()">集群</el-button>
          </div>

        </div>
        <barChart :results="logs" :clickChange="clickChange" @firstClick="getLogData" @secondClick="handleShowLog" />
      </el-tab-pane>
      <el-tab-pane label="事件" name="event" :lazy="true">
        <div class="log_containt_query">
          <!-- 1.简单语句拼写 -->
          <!-- <div class="category">
            分类类别：
          </div> -->
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
    <el-dialog v-model="dialog" :title="title" width="80%" @close="closeDialog" destroy-on-close>
      <logStream v-if="dialog" :log_data="log_stream" :log_total="log_total" v-on:get-more="getMoreLogStream"
        v-on:get-time-range-log="getRangeTimeLog" />
    </el-dialog>
  </div>

</template>

<script setup lang="ts">
import { onMounted, ref, watch, nextTick } from 'vue'
import type { TabsPaneContext } from 'element-plus'
import { Search, Back } from '@element-plus/icons-vue'

import barChart from './barChart.vue';
import logStream from './logStream.vue'
import { useTopoStore } from '@/stores/topo';
import type { logData } from '@/types/index';
import { getELKLogData, getELKProcessLogData, getELKProcessLogStream } from '@/request/elk';
import { calculate_interval } from './utils';
import { ElMessage } from 'element-plus';
import { useConfigStore } from '@/stores/config';
// import result, { host_result, log_query } from './test';

const activeName = ref('log');
const logs = ref([] as logData[]);
const event_logs = ref([] as logData[]);

const dialog = ref(false);
const title = ref('日志详情');
const log_type = ref('log');
const back_btn = ref(false);
const clickChange = ref('');

onMounted(() => {
  getLogData();
})
const handleReset = () => {
  // 页面回到集群
  back_btn.value = false;
  log_type.value = 'log';
  getLogData();
  clickChange.value = 'first';
}

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

// 处理渲染柱状图接口请求体参数
const handleParams = (_params?: any) => {
  let log_query = {
    id: 'log_clusterhost_timeaxis',
    batchId: useConfigStore().topo_request.batch_id,
    params: {
      queryfield_datastream_dataset: "system.syslog",
      queryfield_range_gte: 1719226716185,
      queryfield_range_lte: 1719226836185,
      aggsfield: "agent.hostname",
      size: 0,
      fixed_interval: "5s",
    }
  }
  if (_params) {
    log_query.id = 'log_hostprocess_timeaxis';
    log_query.params.aggsfield = 'process.name';
    Object.assign(log_query.params, { 'queryfield_hostname': _params.seriesName })
    // Object.assign(log_query.params, { 'aggs_1-1_field': 'process.name' })
  }


  switch (log_type.value) {
    case 'log':
      log_query.params.queryfield_range_gte = log_time.value[0].getTime();
      log_query.params.queryfield_range_lte = log_time.value[1].getTime();
      log_query.params.fixed_interval = calculate_interval(log_time.value[0], log_time.value[1]) + 's';
      break;

    default:
      log_query.params.queryfield_range_gte = event_time.value[0].getTime();
      log_query.params.queryfield_range_lte = event_time.value[1].getTime();
      log_query.params.fixed_interval = calculate_interval(event_time.value[0], event_time.value[1]) + 's';
      break;
  }
  return log_query;
}

// 请求日志数据
const bar_params = ref({ hostname: '' } as any); // 记录每次点击柱状图的参数
const getLogData = (_params?: any) => {
  bar_params.value = _params;
  switch (log_type.value) {
    case 'log':
      if (_params || clickChange.value === 'first') {
        // 第一次点击,请求进程信息
        getELKProcessLogData(handleParams(_params)).then(res => {
          if (res.data.code === 200) {
            if (res.data.data.length > 0) {
              logs.value = res.data.data;
              back_btn.value = true;
              clickChange.value = 'second';
            } else {
              ElMessage.info('无数据，请稍后重试')
            }
          } else {
            ElMessage.error(res.data.msg)
          }
        })
      } else {
        // 初始，请求集群信息
        getELKLogData(handleParams()).then(res => {
          if (typeof res === "undefined") {
            ElMessage.error("elk plugin unreachable")
          } else {
            if (res.data.code === 200) {
              if (res.data.data.length > 0) {
                logs.value = res.data.data;
              } else {
                ElMessage.info('无数据，请稍后重试')
              }
            } else {
              ElMessage.error(res.data.msg)
            }
          }
        })
      }

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
      getLogData(bar_params.value);
      break;

    default:
      //
      break;
  }
}

// 请求某一个日志文件流
const log_stream = ref([]);
const logfile_params = ref({} as any);
const log_total = ref(0);
const isRangeLog = ref(false);
const timeRange = ref({} as TimeRange);
const handleShowLog = (_process_info?: any, _size?: number) => {
  logfile_params.value = _process_info;
  if (_process_info) {
    dialog.value = true;
    title.value = _process_info.seriesName + '日志流';
  }
  let log_query = {
    id: "log_stream",
    params: {
      queryfield_datastream_dataset: "system.syslog",
      queryfield_range_gte: _process_info.value[0],
      queryfield_range_lte: _process_info.value[0] + 1000 * 60,
      queryfield_hostname: bar_params.value.hostname ? bar_params.value.hostname : bar_params.value.seriesName,
      queryfield_processname: _process_info.seriesName,
      from: 0,
      size: 20
    }
  }
  if (_size) {
    log_query.params.size = _size;
  }
  if (isRangeLog.value) {
    log_query.params.queryfield_range_gte = timeRange.value.start.getTime()
    log_query.params.queryfield_range_lte = timeRange.value.end.getTime()
  }
  getELKProcessLogStream(log_query).then((res: any) => {
    if (res.data.code === 200) {
      if (res.data.data.hits.length > 0) {
        log_stream.value = res.data.data.hits;
        log_total.value = res.data.data.total;
      } else {
        ElMessage.info('当前时段日志文件无数据')
      }
    } else {
      ElMessage.error(res.data.msg)
    }
  })
}

const getMoreLogStream = (size: number) => {
  handleShowLog(logfile_params.value, size);
}
interface TimeRange {
  start: Date,
  end: Date
}
const getRangeTimeLog = (time_range: TimeRange) => {
  if (time_range) {
    isRangeLog.value = true;
    timeRange.value = time_range;
    handleShowLog(logfile_params.value, 20);
  }
}

// 关闭日志流弹窗事件
const closeDialog = () => {
  isRangeLog.value = false;
  timeRange.value = { start: new Date(new Date().getTime() - 2 * 60 * 60 * 1000), end: new Date() }
  // 清空点击节点信息
  useTopoStore().$reset();
}

</script>


<style scoped lang="scss">
.log_containt {
  width: 100%;
  height: 500px;
  position: absolute;
  bottom: 0;
  right: 0;
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