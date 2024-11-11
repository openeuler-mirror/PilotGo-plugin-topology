<template>
  <div style="width: 100%" v-loading="isloading">
    <div class="search">
      <div class="level">
        选择等级：<el-select
          v-model="level_key"
          placeholder="请选择日志等级"
          style="width: 130px"
          @change="searchLevel"
        >
          <el-option
            v-for="item in level_options"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
      </div>
      &emsp;
      <div class="time">
        时间范围：
        <el-date-picker
          v-model="log_time"
          :disabled="realTime"
          type="datetimerange"
          range-separator="To"
          start-placeholder="Start date"
          end-placeholder="End date"
          @change="ChangeTimeRange"
          @clear="ChangeTimeRange"
        />
      </div>
      &emsp;
      <div class="service">
        选择服务：<el-select
          v-model="service_key"
          placeholder="请选择主机服务"
          style="width: 140px"
          @change="searchService"
        >
          <el-option-group
            v-for="group in service_options"
            :key="group.label"
            :label="group.label"
          >
            <el-option
              v-for="item in group.options"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-option-group>
        </el-select>
      </div>
      &emsp; 
      <div class="level">
        日志模式：<el-select
          v-model="realTime"
          placeholder="请选择日志模式"
          style="width: 120px"
          @change="isResetLog = true"
        >
          <el-option label="实时" :value="true"/>
          <el-option label="非实时" :value="false"/>
        </el-select>
      </div>&emsp;&emsp;
      <el-button type="primary" @click="handleSearch()">查询</el-button>
    </div>
    <div class="log_list">
      <p class="head">
        <span style="width: 200px">时间</span>
        <span style="width: 140px">等级</span>
        <span style="width: calc(100% - 300px)">消息</span>
      </p>
      <ul
        v-infinite-scroll="load"
        :infinite-scroll-distance="1"
        :infinite-scroll-immediate="false"
        class="body"
        style="overflow: auto; height: 470px"
      >
        <li
          v-for="(i, index) in log_stream"
          :title="'&emsp;' + (index + 1) + '.&nbsp;'"
          :name="index"
          :key="index"
        >
          <div style="display: flex; align-items: center">
            <span class="center" style="display: inline-block; width: 200px">{{
              formatDate(new Date(Number(i.timestamp)), "YYYY-MM-DD HH:ii:ss")
            }}</span>
            <span class="center" style="display: inline-block; width: 140px">{{
              levels.find((item) => item.value === i.level)?.label
            }}</span>
            <span
              style="display: inline-block; padding: 0 6px; width: calc(100% - 300px)"
            >
              {{ i.message }}
            </span>
          </div>
        </li>
        <p v-if="loading" style="color: var(--el-color-primary)">
          <el-icon class="is-loading">
            <Loading />
          </el-icon>
          loading...
        </p>
      </ul> 
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watchEffect, watch, onMounted,nextTick, reactive } from "vue";
import { levels } from "./utils";
import { formatDate } from "@/utils/dateFormat";
import { useLogStore } from "@/stores/log";
import { useTopoStore } from'@/stores/topo';

const realTime = ref(false); // 是否实时监听日志变化
const isResetLog = ref(false); // 是否清空日志重新查询
const log_stream = ref([] as logItem[]);
const total_logs = ref(0);
const loading = ref(false);
const isloading = ref(true);
interface logItem {
  timestamp: string;
  message: string;
  level: string;
  targetName: string;
}

let props = defineProps({
  log_data: {
    type: Array as () => logItem[],
    require: false,
    default: [],
  },
  log_total: {
    type: Number,
    default: 0,
    required: false,
  },
  service_list: {
    type: Array as () => SelectGroupItem[],
    default: [],
    required: false,
  },
});

watchEffect(() => {
  if (props.log_data.length > 0) {
    log_stream.value = props.log_data;
    total_logs.value = props.log_total;
    loading.value = false;
    isloading.value = false;
  } else {
    log_stream.value = [];
    total_logs.value = 0;
    isloading.value = false;
  }
});
interface LogSearchObj {
    timeRange:[Date,Date];
    level:string;
    service:{label:string,value:string;}
    realTime:boolean;
}
let logSearchItem = {} as LogSearchObj;
watch([()=> useTopoStore().node_click_info.target_ip,()=> props.service_list],(newVal,oldVal) => {
  if(newVal[0] && newVal[1].length>0) {
    nextTick(() => {
      service_options.value = props.service_list;
      let ip_index = useLogStore().search_list.findIndex(item => item.ip === newVal[0]);
      if(ip_index !== -1) {
        level_key.value = useLogStore().search_list[ip_index].level;
        log_time.value = useLogStore().search_list[ip_index].timeRange;
        service_key.value = useLogStore().search_list[ip_index].service.value;
        realTime.value = useLogStore().search_list[ip_index].realTime;
      } else {
        service_key.value = props.service_list[0].options[0].value;
      }
      handleSearch();
    })
  }
},{immediate:true})

let emit = defineEmits<{
  getWsLogs: [params: {}];
}>();

// 等级搜索功能
const level_key = ref('6');
let level_options = levels;
const searchLevel = (level: string) => {
  isResetLog.value = true;
  logSearchItem.level = level;
};

// 服务搜索功能
interface SelectGroupItem {
  label: string;
  options: [
    {
      value: string;
      label: string;
    }
  ];
}
const service_key = ref('');
let service_options = ref<SelectGroupItem[]>();
const searchService = (service: string) => {
  isResetLog.value = true;
  logSearchItem.service = {
    label:service,
    value:service
  }
};

// 时间筛选功能
const log_time = ref<[Date, Date]>([
  new Date(new Date().getTime() - 2 * 60 * 60 * 1000),
  new Date(),
]);
const ChangeTimeRange = (value: any) => {
  if (value) {
    isResetLog.value = true;
    log_time.value[0] = value[0];
    log_time.value[1] = value[1];
    logSearchItem.timeRange = value;
  }
};

// 查询方法
const handleSearch = () => {
  is_continue.value = true;
  logSearchItem.level = level_key.value;
  logSearchItem.service = {label:service_key.value,value:service_key.value};
  logSearchItem.timeRange = log_time.value;
  logSearchItem.realTime = realTime.value;
  useLogStore().updateLogList({ip:useTopoStore().node_click_info.target_ip,...logSearchItem})
  emit("getWsLogs", {
    severity: level_key.value,
    service: service_key.value,
    timeRange: log_time.value,
    noTail: !realTime.value,
    from: 0,
    size: 20,
    isResetLog: isResetLog.value,
  });
};


const log_size = ref(0);
let is_continue = ref(true);
const load = () => {
  if (total_logs.value == 0 || !is_continue.value || realTime.value) return;
  if (log_size.value >= total_logs.value) {
    log_size.value = total_logs.value;
    is_continue.value = false;
  } else {
    log_size.value = log_size.value + 20;
    is_continue.value = true;
  }
  loading.value = true;
  emit("getWsLogs", {
    severity: level_key.value,
    service: service_key.value,
    timeRange: log_time.value,
    noTail: !realTime.value,
    from: log_size.value,
    size: 20,
    type: 5,
  });
};
</script>

<style scoped lang="scss">
.search {
  height: 44px;
  display: flex;
  align-items: center;
}

.log_list {
  height: 500px;
  width: 100%;
  padding: 0;

  .head {
    margin: 0 1px;
    display: flex;
    align-items: center;
    justify-content: space-around;
    background: var(--el-color-primary-light-9);

    span {
      display: inline-block;
      text-align: center;
    }
  }

  .body {
    margin: 0 1px;
    list-style: none;
    li {
      height: 40px;
    }
  }
}

.border-side {
  border-left: 1px solid var(--el-color-info-light-3);
  border-right: 1px solid var(--el-color-info-light-3);
}

.center {
  text-align: center;
}
</style>
