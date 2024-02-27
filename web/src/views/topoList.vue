<template>
  <my-table ref="confRef" :get-data="getConfList" :del-func="delConfig">
    <template #listName>配置列表</template>
    <template #button_bar>
      <el-button @click="handleDelete">删除</el-button>
    </template>
    <el-table-column type="selection" width="55" />
    <el-table-column prop="id" label="编号" width="80" />
    <el-table-column prop="conf_name" label="配置名称" />
    <el-table-column prop="create_time" label="创建时间" />
    <el-table-column prop="update_time" label="更新时间" />
    <el-table-column prop="description" label="描述" />
    <el-table-column label="操作" width="160">
      <template #default="{ row }">
        <el-button round size="small" @click="handleDetail(row)">详情</el-button>
      </template>
    </el-table-column>
  </my-table>
</template>

<script setup lang="ts">
import { ref } from "vue";
import myTable from '@/components/table.vue';
import { getConfList, delConfig } from "@/request/api";
import { type Config } from "@/types/index";
import { useRouter } from "vue-router";
const router = useRouter();
const confRef = ref();

// 删除
const handleDelete = () => {
  confRef.value.handleDelete();
};
// 查看topo图
const handleDetail = (row: Config) => {
  // router4不支持name+params形式传参，因为刷新丢失问题，用h5的属性history.state中转
  router.push({
    name: 'topoDisplay',
    state: { id: row.id }
  });
};
</script>

<style scoped></style>