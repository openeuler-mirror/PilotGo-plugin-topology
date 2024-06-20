<template>
  <div>
    <el-button @click="clearFilter">重置过滤</el-button>
    <el-table ref="tableRef" row-key="date" :data="tableData" style="width: 100%" height="600px">
      <el-table-column prop="date" label="Date" sortable width="180" column-key="date" />
      <el-table-column prop="name" label="Name" width="180" />
      <el-table-column prop="message" label="message" :formatter="formatter" />
      <el-table-column prop="tag" label="Tag" width="100" :filters="[
      { text: 'topology', value: 'topology' },
      { text: 'elk', value: 'elk' },
    ]" :filter-method="filterTag" filter-placement="bottom-end">
        <template #default="scope">
          <el-tag :type="scope.row.tag === 'topology' ? 'primary' : 'success'" disable-transitions>{{ scope.row.tag
            }}</el-tag>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { TableColumnCtx, TableInstance } from 'element-plus'

let props = defineProps({
  log_data: {
    type: Object,
    require: false,
    default: {}
  }
})

let emit = defineEmits<{
  removeKey: [value: string]
}>()

interface Log {
  date: string
  name: string
  message: string
  tag: string
}

const tableRef = ref<TableInstance>()

const clearFilter = () => {
  tableRef.value!.clearFilter()
}
const formatter = (row: Log, _column: TableColumnCtx<Log>) => {
  return row.message
}
const filterTag = (value: string, row: Log) => {
  return row.tag === value
}

const tableData: Log[] = [
  {
    date: '2016-05-03',
    name: 'test',
    message: 'No. 189, Grove St, Los Angeles',
    tag: 'topology',
  },
]

</script>

<style scoped></style>