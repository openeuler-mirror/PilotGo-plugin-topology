<template>
  <div style="width: 100%" v-loading="isloading">
    <div class="search">
      <div class="level">
        选择等级：<el-select
          v-model="level_key"
          placeholder="请选择日志等级"
          style="width: 180px"
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
          style="width: 180px"
          @change="searchService"
        >
          <!-- <el-option v-for="item in service_options" :key="item.value" :label="item.label" :value="item.value" /> -->
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
      &emsp; <el-button type="primary" @click="handleSearch()">查询</el-button>&emsp;
      实时：<el-switch v-model="realTime" @change="openRealTimeLog" />
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
        style="overflow: auto; height: 350px"
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
import { ref, watchEffect, watch, computed, onMounted } from "vue";
import { levels } from "./utils";
import { formatDate } from "@/utils/dateFormat";

const realTime = ref(false); // 是否实时监听日志变化
const isResetLog = ref(false); // 是否清空日志重新查询
const log_stream = ref([] as logItem[]);
const total_logs = ref(0);
const activeNames = ref([0]);
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
watch(
  () => props.service_list,
  (new_list) => {
    if (new_list.length > 0) {
      service_options.value = props.service_list;
      console.log(props.service_list);
      service_key.value = props.service_list[0].options[0].value;
      handleSearch();
    }
  }
);
watchEffect(() => {
  if (props.log_data.length > 0) {
    log_stream.value = props.log_data;
    total_logs.value = props.log_total;
    loading.value = false;
    isloading.value = false;
    for (let i = 0; i < props.log_data.length; i++) {
      activeNames.value.push(i);
    }
  } else {
    log_stream.value = [];
    total_logs.value = 0;
    isloading.value = false;
  }
});
interface TimeRange {
  start: Date;
  end: Date;
}
let emit = defineEmits<{
  getMore: [value: number];
  getTimeRangeLog: [time_range: TimeRange];
  getWsLogs: [params: {}];
}>();

// 等级搜索功能
const level_key = ref("6");
let level_options = levels;
const searchLevel = (level: string) => {
  isResetLog.value = true;
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
const service_key = ref("");
let service_options = ref<SelectGroupItem[]>();
const searchService = (service: string) => {
  // console.log(service)
  isResetLog.value = true;
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
  }
};

// 查询方法
const handleSearch = () => {
  emit("getWsLogs", {
    severity: level_key.value,
    service: service_key.value,
    timeRange: log_time.value,
    noTail: true,
    from: 0,
    size: 20,
    isResetLog: isResetLog.value,
  });
};

// 实时监听日志功能
const openRealTimeLog = (state: boolean) => {
  is_continue.value = !state;
  emit("getWsLogs", {
    severity: level_key.value,
    service: service_key.value,
    timeRange: state ? ["", ""] : log_time.value,
    noTail: !state,
  });
};

const log_size = ref(0);
let is_continue = ref(true);
const load = () => {
  if (total_logs.value !== 0 && !is_continue.value) return;
  console.log("滑动到底部可发起请求", total_logs.value, log_size.value);
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
    noTail: true,
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
  height: 400px;
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
      height: 50px;
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
