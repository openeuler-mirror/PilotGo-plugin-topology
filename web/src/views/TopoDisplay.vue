<!--
 * Copyright (c) KylinSoft  Co., Ltd. 2024.All rights reserved.
 * PilotGo-plugin-topology licensed under the Mulan Permissive Software License, Version 2. 
 * See LICENSE file for more details.
 * Author: zhaozhenfang <zhaozhenfang@kylinos.cn>
 * Date: Mon Mar 4 17:27:28 2024 +0800
-->
<template>
  <div id="topoDisplay" class="topoContaint" v-loading="loading">
    <el-page-header @back="goBack">
      <template #content>
        <span style="font-size: 14px"> 拓扑图 </span>
      </template>
      <template #extra>
        <div class="flex items-center">
          <el-dropdown @command="changeInterval">
            <span style="font-size: 14px"> 定时器 </span>
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
    <PGTopo
      ref="topoRef"
      style="width: 100%; height: calc(100% - 50px)"
      :graph_mode="graphMode"
      :time_interval="timeInterval"
      v-on:click-topo-canvas="clickTopoCanvas"
    />
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
    <el-dialog
      v-model="dialog"
      :title="title"
      width="86%"
      @close="closeDialog"
      destroy-on-close
    >
      <logStream
        v-if="log_type === 'elk'"
        :log_data="elklog_stream"
        :log_total="elklog_total"
        v-on:get-more="getMoreLogStream"
        v-on:get-time-range-log="getRangeTimeLog"
      />

      <logStream
        v-show="log_type === 'plugin'"
        :log_data="pluginlog_stream"
        :log_total="pluginlog_total"
        :service_list="serviceList"
        v-on:get-ws-logs="getWsLogs"
      />
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ElMessage } from "element-plus";
import { onBeforeUnmount, onMounted, ref, watch, watchEffect, nextTick } from "vue";
import PGTopo from "@/components/PGTopo.vue";
import nodeDetail from "./nodeDetail.vue";
import LogChart from "./topoLogs/index.vue";
import logStream from "./topoLogs/logStream_plugin.vue";

import { getCustomTopo, getTopoData, getUuidTopo } from "@/request/api";
import { getELKProcessLogStream } from "@/request/elk";
import { useTopoStore } from "@/stores/topo";
import { useConfigStore } from "@/stores/config";
import router from "@/router";
import socket from "@/utils/socket";
import { formatDate } from "@/utils/dateFormat";
import { useLogStore } from "@/stores/log";

const graphMode = ref("default");
const timeInterval = ref("关闭");
let time_interval_num: number = 0;
let timer: any;
const showTopo = ref(false);
const loading = ref(false);
let request_type: string;
let request_id: string | number;

let showLogChart = ref(false);
const logChart = ref(null);

const topoRef = ref(null);

const dialog = ref(false);
const title = ref("日志详情");
const log_type = ref("plugin");
onMounted(() => {
  loading.value = true;
});
onBeforeUnmount(() => {
  // 离开页面，清空点击事件缓存数据
  useTopoStore().$reset();
  useLogStore().$reset();
});

// 点击画布关闭日志tab
const clickTopoCanvas = () => {
  showLogChart.value = false;
};

const goBack = () => {
  if (timer) {
    clearInterval(timer);
  }
  router.push("topoList");
};

watchEffect(() => {
  request_type = useConfigStore().topo_request.type;
  request_id = useConfigStore().topo_request.id;
  nextTick(() => {
    let topoData = {};
    switch (request_type) {
      case "custom":
        getCustomTopo({ id: request_id as number }).then((res) => {
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
            router.push("topoList");
          }
        });

        topoTimer(request_id, "关闭");

        break;
      case "single":
        getUuidTopo({ uuid: request_id as string }).then((res) => {
          if (res.data.code === 200) {
            topoData = res.data.data;
            loading.value = false;
            showTopo.value = true;
            setTimeout(() => {
              useTopoStore().topo_data = topoData;
            }, 200);
          } else {
            ElMessage.error(res.data.msg);
            router.push("topoList");
          }
        });
        break;

      default:
        getTopoData().then((res) => {
          if (res.data.code === 200) {
            topoData = res.data.data;
            useTopoStore().topo_data = topoData;
            showTopo.value = true;
            loading.value = false;
          }
        });
        break;
    }
  });
});

watch(
  () => timeInterval.value,
  (newdata) => {
    topoTimer(request_id, newdata);
  }
);

function changeInterval(command: string) {
  timeInterval.value = command;
}

// 定时器
function topoTimer(request_id: any, interval: string) {
  let topoData = {};

  if (interval === "关闭") {
    clearInterval(timer);
  } else {
    switch (timeInterval.value) {
      case "5s":
        time_interval_num = 5000;
        break;
      case "10s":
        time_interval_num = 10000;
        break;
      case "15s":
        time_interval_num = 15000;
        break;
      case "1m":
        time_interval_num = 60000;
        break;
      case "5m":
        time_interval_num = 300000;
        break;
    }

    try {
      if (timer) {
        clearInterval(timer);
      }
      timer = setInterval(() => {
        getCustomTopo({ id: request_id as number }).then((res) => {
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
            router.push("topoList");
          }
        });
      }, time_interval_num);
    } catch (error) {
      console.error(error);
    }
  }
}

// 请求某一个日志文件流
interface TimeRange {
  start: Date;
  end: Date;
}
const elklog_stream = ref([]);
const logfile_params = ref({} as any);
const elklog_total = ref(0);
const isRangeLog = ref(false);
const timeRange = ref({} as TimeRange);
const handleShowLog = (node_info?: any, _size?: number) => {
  logfile_params.value = node_info;
  if (node_info) {
    dialog.value = true;
    title.value = node_info.process_name + "日志流";
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
      size: 20,
    },
  };
  if (_size) {
    log_query.params.size = _size;
  }
  if (isRangeLog.value) {
    log_query.params.queryfield_range_gte = timeRange.value.start.getTime();
    log_query.params.queryfield_range_lte = timeRange.value.end.getTime();
  }
  getELKProcessLogStream(log_query).then((res: any) => {
    if (res.data.code === 200) {
      if (res.data.data.hits.length > 0) {
        elklog_stream.value = res.data.data.hits;
        elklog_total.value = res.data.data.total;
      } else {
        ElMessage.info("当前文件无日志数据");
      }
    } else {
      ElMessage.error(res.data.msg);
    }
  });
};

const getMoreLogStream = (size: number) => {
  handleShowLog(logfile_params.value, size);
};

const getRangeTimeLog = (time_range: TimeRange) => {
  if (time_range) {
    isRangeLog.value = true;
    timeRange.value = time_range;
    handleShowLog(logfile_params.value, 20);
  }
};

// ----------与目标服务器通信消息处理-----------
interface logItem {
  timestamp: string;
  message: string;
  level: string;
  targetName: string;
}
const serviceList = ref([] as any);
const pluginlog_stream = ref([] as logItem[]);
const pluginlog_total = ref(0);

watch(()=>useLogStore().ws_isOpen,newV => {
  if(newV) {
    socket.send({
      type: 1,
      joptions: null,
      data: useTopoStore().node_click_info.target_ip + ":9995",
    });
  }
},{immediate:true})

const receiveMessage = (message: any) => {
  let result = JSON.parse(message.data);
  switch (result.type) {
    case 3:
      // 与目标机器建立连接
      socket.send({ type: 2, joptions: null, data: null });
      break;

    default:
      if (result.data.type === 0) {
        // 返回消息属于日志条目
        if (noTail.value) {
          // 固定时间段
          pluginlog_stream.value = pluginlog_stream.value.concat(result.data.data.hits);
          pluginlog_total.value = result.data.data.total;
        } else {
          pluginlog_stream.value.push(result.data.data);
        }
      } else {
        // 返回消息属于主机服务列表
        let severiceOptios = [] as any[];
        Object.keys(result.data.data).forEach((i: string) => {
          let selectItem = { label: "", options: [] };
          selectItem["label"] = i;
          selectItem["options"] = result.data.data[i].map((j: string) => {
            let opt = { label: "", value: "" };
            opt["label"] = j;
            opt["value"] = j;
            return opt;
          });
          severiceOptios.push(selectItem);
        });
        serviceList.value = JSON.parse(JSON.stringify(severiceOptios));
      }
      break;
  }
};
// 发送ws日志请求
const noTail = ref(false); // false:实时  true:固定时间段
const getWsLogs = (params: any) => {
  noTail.value = params.noTail;
  if (params.isResetLog) {
    pluginlog_stream.value = [];
    pluginlog_total.value = 0;
  }
  let joptions = {
    severity: params.severity,
    since:
      params.timeRange[0] == ""
        ? ""
        : formatDate(params.timeRange[0], "YYYY-MM-DD HH:ii:ss"),
    until:
      params.timeRange[1] == ""
        ? ""
        : formatDate(params.timeRange[1], "YYYY-MM-DD HH:ii:ss"),
    unit: "",
    user: "",
    transport: "",
    notail: params.noTail,
    from: params.from,
    size: params.size ? params.size : null,
  } as any;
  if(!serviceList.value) return;
  let selected_service = serviceList.value.find((group: any) =>
    group.options.some((option: any) => option.label == params.service)
  );
  selected_service.label === "systemd"
    ? (joptions["unit"] = params.service)
    : (joptions[`${selected_service.label}`] = params.service);
  socket.send({
    type: params.type ? params.type : 0,
    joptions,
    data: null,
  });
};

// 关闭日志流弹窗事件
const closeDialog = () => {
  isRangeLog.value = false;
  timeRange.value = {
    start: new Date(new Date().getTime() - 2 * 60 * 60 * 1000),
    end: new Date(),
  };
  useTopoStore().$reset();
  socket.close();
  pluginlog_stream.value = [];
  pluginlog_total.value = 0;
  serviceList.value = [];
};

/*
 * 监听topo图节点右键查看日志事件
 * node_click_info:{host_name:节点|父节点name,node_id:节点id,process_name:进程name}
 */
watch(
  () => useTopoStore().node_click_info,
  (node_click_info) => {
    if (node_click_info.node_id) {
      dialog.value = true;
      if (log_type.value === "elk") {
        let query_params = {
          process_name: node_click_info.process_name,
          value: [new Date().getTime() - 60 * 1000],
          host_name: node_click_info.host_name,
        };
        handleShowLog(query_params);
      } else {
        socket.init(receiveMessage, "");
      }
    }
  },
  { immediate: true, deep: true }
);
</script>

<style scoped lang="scss">
.topoContaint {
  width: 96%;
  height: 100%;
  margin: 0 auto;
  position: relative;
  // 设置文字双击不能选中
  -webkit-user-select: none; /* Safari */
  -moz-user-select: none;    /* Firefox */
  -ms-user-select: none;     /* IE/Edge */
  user-select: none;         /* 标准语法 */

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
    transition: all 0.5s ease-in-out;
  }

  .fade-enter-from,
  .fade-leave-to {
    transform: translateY(100%);
  }
}
</style>
