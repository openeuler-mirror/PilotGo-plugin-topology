<!--
 * Copyright (c) KylinSoft  Co., Ltd. 2024.All rights reserved.
 * PilotGo-plugin-topology licensed under the Mulan Permissive Software License, Version 2. 
 * See LICENSE file for more details.
 * Author: zhaozhenfang <zhaozhenfang@kylinos.cn>
 * Date: Wed Jul 17 16:38:05 2024 +0800
-->
<template>
  <el-table :data="tableData" stripe class="table" :cell-style="{ borderBottom: '1px solid #999 !important' }">
    <el-table-column v-for="(col, index) in columnList" :prop="col!.prop" :label="col!.label" :key="index"
      style="background-color: red;" />
  </el-table>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue';
interface tableItem {
  prop: string,
  label: string
}

const props = defineProps({
  tableData: {
    type: Array,
    default: [],
    required: true
  },
  columnList: {
    type: Array,
    default: [],
    required: true
  },
})

const tableData = ref([]);
const columnList = ref<tableItem[]>([]);

onMounted(() => {
  tableData.value = props.tableData as any;
  columnList.value = props.columnList as tableItem[];
})

watch(() => props.tableData, (new_data) => {
  tableData.value = new_data as any;
}, {
  deep: true
})

watch(() => props.columnList, (new_cols) => {
  columnList.value = new_cols as tableItem[];
}, {
  deep: true
})

</script>

<style scoped lang="scss">
.table {
  width: 100%;
}
</style>