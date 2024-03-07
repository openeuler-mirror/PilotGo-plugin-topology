<template>
  <div>
    <el-drawer v-model="drawer" direction="ttb" size="300px">
      <template #header>
        <h4>请选择你想查看的主机范围</h4>
      </template>
      <template #default>
        <div class="dra_content">
          <el-radio-group v-model="select_type">
            <el-radio label="single" size="large">单机</el-radio>
            <el-radio label="multi" size="large">多机</el-radio>
          </el-radio-group>
          <div class="dra_content_select" v-show="select_type === 'single'">
            <el-form :inline="true" :model="formInline" class="demo-form-inline">
              <el-form-item label="批次">
                <el-select v-model="formInline.batchId" placeholder="请选择批次" clearable @change="handlebatchDetail">
                  <el-option v-for="item in batchs" :label="item.name" :value="item.id" />
                </el-select>
              </el-form-item>
              <el-form-item label="主机">
                <el-select v-model="formInline.uuid" placeholder="请选择主机" clearable>
                  <el-option v-for="item in hosts" :label="item.ip" :value="item.uuid" />
                </el-select>
              </el-form-item>
            </el-form>
          </div>
        </div>
      </template>
      <template #footer>
        <div style="flex: auto">
          <el-button @click="handleClose">取消</el-button>
          <el-button type="primary" @click="handleConfirm">确认</el-button>
        </div>
      </template>
    </el-drawer>
    <my-table ref="confRef" :get-data="getConfList" :del-func="delConfig">
      <template #listName>配置列表</template>
      <template #button_bar>
        <el-button @click="drawer = true">单/多机</el-button>
        <el-button @click="handleDelete">删除</el-button>
      </template>
      <el-table-column type="selection" width="55" />
      <el-table-column prop="id" label="编号" width="80" />
      <el-table-column prop="conf_name" label="配置名称" />
      <el-table-column prop="conf_version" label="版本" />
      <el-table-column prop="create_time" label="创建时间" />
      <el-table-column prop="update_time" label="更新时间" />
      <el-table-column prop="description" label="描述" />
      <el-table-column label="操作" width="240">
        <template #default="{ row }">
          <el-button round size="small" @click="handleDetail(row)">拓扑图</el-button>
          <el-button round size="small" @click="handleConfig(row)">json配置</el-button>
          <el-button round size="small" @click="handleEdit(row)">编辑</el-button>
        </template>
      </el-table-column>
    </my-table>
    <!-- 配置json展示 -->
    <el-dialog v-model="showDialog" title="配置详情" width="800">
      <el-scrollbar height="600"> 
        <vue-json-pretty :data="configJson" showLength/>
      </el-scrollbar>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from "vue";
import myTable from '@/components/table.vue';
import { getConfList, delConfig, getBatchList, getBatchDetail } from "@/request/api";
import type{  Config, TopoCustomFormType } from "@/types/index";
import { useRouter } from "vue-router";
import { useConfigStore } from "@/stores/config";
import VueJsonPretty from 'vue-json-pretty';
import 'vue-json-pretty/lib/styles.css';
const router = useRouter();
const confRef = ref();
const drawer = ref(false);
const select_type = ref('single');
const showDialog = ref(false);
let configJson = reactive<TopoCustomFormType>;
const formInline = reactive({
  batchId: null,
  uuid: '',
})

let batchs = reactive([
  {
    id: null,
    name: ''
  }
]);
let hosts = reactive([
  {
    uuid: '',
    ip: ''
  }
])
onMounted(() => {
  getBatchList().then(res => {
    if (res.data.code == 200) {
      batchs = res.data.data;
    }
  })
})

// 选择主机
const handlebatchDetail = (batchId: number) => {
  getBatchDetail({ batchId: batchId }).then(res => {
    if (res.data.code === 200) {
      hosts = res.data.data;
    }
  })
}

// 查看topo配置详情
const handleConfig = (row:any) => {
  showDialog.value = true;
  configJson = row;
}

// 编辑配置文件
const handleEdit = (row:any) => {
  useConfigStore().topo_config = row;
  router.push('customTopo');
}

// 关闭抽屉
const handleClose = () => {
  drawer.value = false;
}
// 选中
const handleConfirm = () => {
  router.push({
    name: 'topoDisplay',
    state: { type: select_type.value, id: formInline.uuid }
  });
}
// 删除
const handleDelete = () => {
  confRef.value.handleDelete();
};
// 查看topo图
const handleDetail = (row: Config) => {
  router.push({
    name: 'topoDisplay',
    state: { type: 'custom', id: row.id }
  });
};
</script>

<style scoped></style>