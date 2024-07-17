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