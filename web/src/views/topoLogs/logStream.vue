<!--
 * Copyright (c) KylinSoft  Co., Ltd. 2024.All rights reserved.
 * PilotGo-plugin-topology licensed under the Mulan Permissive Software License, Version 2. 
 * See LICENSE file for more details.
 * Author: zhaozhenfang <zhaozhenfang@kylinos.cn>
 * Date: Thu Jun 20 17:31:08 2024 +0800
-->
<template>
  <div style="width: 100%;" v-loading="isloading">
    <div class="search">
      <div class="level">
        选择等级：<el-select v-model="level_key" placeholder="请选择日志等级" style="width: 240px" @change="searchLevel">
          <el-option v-for="item in level_options" :key="item.value" :label="item.label" :value="item.value" />
        </el-select>
      </div>&emsp;
      <div class="time">
        时间范围：
        <el-date-picker v-model="log_time" type="datetimerange" range-separator="To" start-placeholder="Start date"
          end-placeholder="End date" @change="ChangeTimeRange" @clear="ChangeTimeRange" />
      </div>
    </div>
    <div class="infinite-list">
      <p class="head">
        <span style=" width: 200px;">时间</span>
        <span style="width:100px;">等级</span>
        <span style="width:calc(100% - 300px);">消息</span>
      </p>
      <el-collapse v-infinite-scroll="load" v-model="activeNames" class="body" style="overflow: auto; height:550px;">
        <el-collapse-item v-for="(i, index) in log_stream" :title="'&emsp;' + (index + 1) + '.&nbsp;' + i.date"
          :name="index" :key="index">
          <div style="display: flex;align-items: center">
            <span class="center" style="display: inline-block; width: 200px;">{{ i.date }}</span>
            <span class="center" style="display: inline-block; width:100px;">{{ i.level === "" ? '暂无' : i.level
              }}</span>
            <span style="display: inline-block;padding:0 6px;width:calc(100% - 300px);">
              {{ i.message }}
            </span>
          </div>
        </el-collapse-item>
        <p v-if="loading" style="color:var(--el-color-primary)">
          <el-icon class="is-loading">
            <Loading />
          </el-icon>
          loading...
        </p>
      </el-collapse>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watchEffect, computed, onMounted } from 'vue'
import { levels } from './utils'

const log_stream = ref([] as logItem[]);
const total_logs = ref(0);
const activeNames = ref([0])
const loading = ref(false);
const isloading = ref(true);
interface logItem {
  date: string;
  message: string;
  level: string;
}
let props = defineProps({
  log_data: {
    type: Array as () => logItem[],
    require: false,
    default: []
  },
  log_total: {
    type: Number,
    default: 0,
    required: true
  }
})
onMounted(() => {
  isloading.value = true;
  if (props.log_data.length == 0) {
    isloading.value = false;
  }
})
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
  }
})
interface TimeRange {
  start: Date,
  end: Date
}
let emit = defineEmits<{
  getMore: [value: number],
  getTimeRangeLog: [time_range: TimeRange]
}>()

// 等级搜索功能
const level_key = ref('');
let level_options = levels;
const searchLevel = (level: string) => {
  console.log(level)
}

// 时间筛选功能
const log_time = ref<[Date, Date]>([
  new Date(new Date().getTime() - 2 * 60 * 60 * 1000),
  new Date(),
])
const ChangeTimeRange = (value: any) => {
  if (value) {
    log_time.value[0] = value[0]
    log_time.value[1] = value[1]
    emit('getTimeRangeLog', { start: value[0], end: value[1] });
  }
}

const log_size = ref(20);
let is_continue = ref(true);
const load = () => {
  if (!total_logs.value || !is_continue.value) return;
  if (log_size.value >= total_logs.value) {
    log_size.value = total_logs.value;
    is_continue.value = false;
  } else {
    log_size.value = log_size.value + 20;
    is_continue.value = true;
  }
  loading.value = true;
  emit('getMore', log_size.value);
}
</script>

<style scoped lang="scss">
.search {
  height: 44px;
  display: flex;
  align-items: center;

}

.infinite-list {
  height: 600px;
  width: 100%;
  padding: 0;
  list-style: none;

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