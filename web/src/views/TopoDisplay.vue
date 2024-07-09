<template>

  <div id="topoDisplay" class="topoContaint" v-loading="loading">
    <el-page-header @back="goBack">
      <template #content>
        <span style="font-size: 14px;"> 拓扑图 </span>
      </template>
      <template #extra>
        <div class="flex items-center">
          <el-dropdown @command="changeInterval">
            <span style="font-size: 14px;"> 定时器 </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="关闭">关闭</el-dropdown-item>
                <el-dropdown-item command="5s">5s</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </template>
    </el-page-header>
    <!-- 展示topo图 -->
    <PGTopo ref="topoRef" style="width: 100%;height: calc(100% - 50px);" :graph_mode="graphMode"
      :time_interval="timeInterval" v-on:click-topo-canvas="clickTopoCanvas" />
    <!-- 嵌套抽屉组件展示数据 -->
    <nodeDetail />
    <!-- 嵌套抽屉组件展示日志信息 -->
    <Transition name="fade">
      <LogChart v-show="showLogChart" ref="logChart" />
    </Transition>
    <el-icon class="upLog" @click="showLogChart = !showLogChart">
      <ArrowUpBold class="up_log_up" v-if="!showLogChart" />
      <ArrowDownBold class="up_log_down" v-else />
    </el-icon>
    <el-dialog v-model="dialog" :title="title" width="80%" @close="closeDialog" destroy-on-close>
      <logStream v-if="dialog" :log_data="log_stream" :log_total="log_total" v-on:get-more="getMoreLogStream"
        v-on:get-time-range-log="getRangeTimeLog" />
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ElMessage } from 'element-plus'
import { onBeforeUnmount, onMounted, ref, watch, watchEffect, nextTick } from 'vue';
import PGTopo from '@/components/PGTopo.vue';
import nodeDetail from './nodeDetail.vue';
import LogChart from './topoLogs/index.vue';
import logStream from './topoLogs/logStream.vue';

import { getCustomTopo, getTopoData, getUuidTopo } from "@/request/api";
import { getELKProcessLogStream } from '@/request/elk';
import { useTopoStore } from '@/stores/topo';
import { useConfigStore } from '@/stores/config';
import router from '@/router';

const graphMode = ref('default');
const timeInterval = ref('关闭');
let time_interval_num: number = 0
let timer: any;
const showTopo = ref(false);
const loading = ref(false);
let request_type: string
let request_id: string | number

let showLogChart = ref(false);
const logChart = ref(null)

const topoRef = ref(null)

const dialog = ref(false);
const title = ref('日志详情');
onMounted(() => {
  loading.value = true;
})
onBeforeUnmount(() => {
  // 离开页面，清空点击事件缓存数据
  useTopoStore().$reset();
})

// 点击画布关闭日志tab
const clickTopoCanvas = (_e: any) => {
  showLogChart.value = false;
}


const goBack = () => {
  if (timer) {
    clearInterval(timer);
  }
  router.push('topoList');
}

watchEffect(() => {
  request_type = useConfigStore().topo_request.type;
  request_id = useConfigStore().topo_request.id;
  nextTick(() => {
    let topoData = {};
    switch (request_type) {
      case 'custom':
        getCustomTopo({ id: request_id as number }).then(res => {
          if (res.data.code === 200) {
            topoData = res.data.data;
            loading.value = false;
            showTopo.value = true;
            // setTimeout(() => {
            //   useTopoStore().topo_data = topoData;
            // }, 200)
            useTopoStore().topo_data = topoData;
          } else {
            ElMessage.error(res.data.msg);
            router.push('topoList');
          }
        })

        topoTimer(request_id, '关闭');

        break;
      case 'single':
        getUuidTopo({ uuid: request_id as string }).then(res => {
          if (res.data.code === 200) {
            topoData = res.data.data;
            loading.value = false;
            showTopo.value = true;
            setTimeout(() => {
              useTopoStore().topo_data = topoData;
            }, 200)
          } else {
            ElMessage.error(res.data.msg);
            router.push('topoList');
          }
        })
        break;

      default:
        getTopoData().then(res => {
          if (res.data.code === 200) {
            topoData = res.data.data;
            useTopoStore().topo_data = topoData;
            showTopo.value = true;
            loading.value = false;
          }
        })
        break;
    }


  })

})

watch(() => timeInterval.value, (newdata) => {
  topoTimer(request_id, newdata);
})

function changeInterval(command: string) {
  timeInterval.value = command;
}

// 定时器
function topoTimer(request_id: any, interval: string) {
  let topoData = {};

  if (interval === '关闭') {
    clearInterval(timer)
  } else {
    switch (timeInterval.value) {
      case '5s':
        time_interval_num = 5000
        break;
      case '10s':
        time_interval_num = 10000
        break;
      case '15s':
        time_interval_num = 15000
        break;
      case '1m':
        time_interval_num = 60000
        break;
      case '5m':
        time_interval_num = 300000
        break;
    }

    try {
      if (timer) {
        clearInterval(timer)
      }
      timer = setInterval(() => {
        getCustomTopo({ id: request_id as number }).then(res => {
          if (res.data.code === 200) {
            topoData = res.data.data;
            loading.value = false;
            showTopo.value = true;
            // setTimeout(() => {
            //   useTopoStore().topo_data = topoData;
            // }, 200)
            useTopoStore().topo_data = topoData;
          } else {
            ElMessage.error(res.data.msg);
            router.push('topoList');
          }
        })
      }, time_interval_num)

    } catch (error) {
      console.error(error)
    }
  }
}

// 请求某一个日志文件流
interface TimeRange {
  start: Date,
  end: Date
}
const log_stream = ref([]);
const logfile_params = ref({} as any);
const log_total = ref(0);
const isRangeLog = ref(false);
const timeRange = ref({} as TimeRange);
const handleShowLog = (node_info?: any, _size?: number) => {
  logfile_params.value = node_info;
  if (node_info) {
    dialog.value = true;
    title.value = node_info.process_name + '日志流';
  }
  let log_query = {
    id: "log_stream",
    params: {
      queryfield_datastream_dataset: "application",
      queryfield_range_gte: new Date().getTime() - 2 * 60 * 60 * 1000,
      queryfield_range_lte: new Date().getTime(),
      queryfield_hostname: node_info.host_name,
      queryfield_processname: node_info.process_name,
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
        ElMessage.info('当前文件无日志数据')
      }
    } else {
      ElMessage.error(res.data.msg)
    }
  })
}

const getMoreLogStream = (size: number) => {
  handleShowLog(logfile_params.value, size);
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
}

/* 
* 监听topo图节点右键查看日志事件
* node_click_info:{host_name:节点|父节点name,node_id:节点id,process_name:进程name}
*/
watch(() => useTopoStore().node_click_info, (node_click_info) => {
  if (node_click_info.node_id) {
    dialog.value = true;
    let query_params = {
      process_name: node_click_info.process_name,
      value: [new Date().getTime() - 60 * 1000],
      host_name: node_click_info.host_name,
    };
    handleShowLog(query_params);
  }

}, { immediate: true, deep: true })
</script>

<style scoped lang="scss">
.topoContaint {
  width: 96%;
  height: 100%;
  margin: 0 auto;
  position: relative;

  @keyframes bounce {

    0%,
    100% {
      transform: translateY(0);
    }

    50% {
      transform: translateY(-10px);
    }
  }

  .upLog {
    position: relative;
    left: 50%;
    cursor: pointer;
    font-size: 20px;
    animation: bounce 2s infinite;

    &:hover {
      color: #409efc;
    }
  }

  .fade-enter-active,
  .fade-leave-active {
    transition: all .5s ease-in-out;
  }

  .fade-enter-from,
  .fade-leave-to {
    transform: translateY(100%);
  }
}
</style>