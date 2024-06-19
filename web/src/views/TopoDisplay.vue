<template>

  <div class="topoContaint" v-loading="loading">
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
    <PGTopo style="width: 100%;height: calc(100% - 50px);" :graph_mode="graphMode" :time_interval="timeInterval" />
    <!-- 嵌套抽屉组件展示数据 -->
    <nodeDetail />
    <!-- 嵌套抽屉组件展示日志信息 -->
    <Transition name="fade">
      <LogChart v-if="showLogChart" />
    </Transition>
    <el-icon class="upLog" @click="showLogChart = !showLogChart">
      <ArrowUpBold class="up_log_up" v-if="!showLogChart" />
      <ArrowDownBold class="up_log_down" v-else />
    </el-icon>

  </div>
</template>

<script setup lang="ts">
import { ElMessage } from 'element-plus'
import { onMounted, ref, watch, watchEffect, nextTick } from 'vue';
import PGTopo from '@/components/PGTopo.vue';
import nodeDetail from './nodeDetail.vue';
import LogChart from './topoLogs/index.vue';

import { getCustomTopo, getTopoData, getUuidTopo } from "@/request/api";
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

onMounted(() => {
  loading.value = true;
})

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

watch(() => useTopoStore().node_log_id, (new_log_id) => {
  if (new_log_id) {
    showLogChart.value = true;
  }
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

</script>

<style scoped lang="scss">
.topoContaint {
  width: 96%;
  height: 100%;
  margin: 0 auto;

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