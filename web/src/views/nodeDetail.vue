<!-- 抽屉组件展示节点详情 -->
<template>
  <!-- 外层抽屉组件 -->
  <el-drawer class="drawer" v-model="display_drawer" :with-header="false" direction="rtl" size="560px"
    :before-close="handleClose" :modal="true" :append-to-body="false">
    <div class="drawer_top_div">
      <el-descriptions title="节点信息" :column="1" class="info">
        <el-descriptions-item label="节点名称：">{{ node.name }}</el-descriptions-item>
        <el-descriptions-item label="节点类型：">{{ node.type }}</el-descriptions-item>
        <el-descriptions-item label="节点tags：">
          <el-scrollbar height="100px">
            <el-tag v-for="(tag, i) in tags" :key="i" effect="plain">{{
              tag
            }}</el-tag>
          </el-scrollbar>
        </el-descriptions-item>
      </el-descriptions>

      <el-button-group class="button">
        <el-tooltip v-for="(btn, index) in btns" :content="btn.label" placement="bottom" effect="light">
          <el-button :icon="handleIcon(index)" color="#79bbff" @click="handleBtnClick(btn)" :disabled="btn.disabled" plain
            circle />
        </el-tooltip>
      </el-button-group>
    </div>

    <!-- 图表 -->
    <div class="drawer_body_div" v-if="isHost">
      <span>时间范围：</span>
      <el-date-picker v-model="dateRange" type="datetimerange" :shortcuts="pickerOptions" range-separator="至"
        start-placeholder="开始日期" end-placeholder="结束日期" @change="changeDate">
      </el-date-picker>
      <grid-layout :col-num="3" :is-draggable="grid.draggable" :is-resizable="grid.resizable" :layout.sync="layout"
        :row-height="100" :use-css-transforms="true" :vertical-compact="true" :responsive="true">
        <template v-for="(item, i) in layout">
          <grid-item :key="i" :h="item.h" :i="item.i" :static="item.static" :w="item.w" :x="item.x" :y="item.y" :min-w="2"
            :min-h="2" @resize="SizeAutoChange(item.i, item.query.isChart)" @resized="SizeAutoChange"
            drag-allow-from=".drag" drag-ignore-from=".noDrag" v-if="item.display">
            <div class="drag">
              <span class="drag-title">{{ item.title }}</span>
            </div>
            <div class="noDrag">
              <MyEcharts :query="item.query" :startTime="startTime" :endTime="endTime"
                :style="{ 'width': '100%', 'height': '100%' }">
              </MyEcharts>
            </div>
          </grid-item>
        </template>
      </grid-layout>
    </div>

    <!-- 嵌套抽屉组件 -->
    <div class="inner_drawer_div">
      <el-drawer v-model="inner_drawer.display" :title="inner_drawer.label" :append-to-body="true" size="380px">
        <el-table :data="table_data" stripe style="width: 100%" v-if="inner_drawer.type === 'metric'">
          <el-table-column prop="name" label="属性" />
          <el-table-column prop="value" label="值" />
        </el-table>
        <el-checkbox v-if="inner_drawer.type === 'chart'" v-for="item in layout" v-model="item.display"
          :label="item.title" size="large" />
      </el-drawer>
    </div>
  </el-drawer>
</template>

<script setup lang="ts">
import { ref, reactive, watch } from "vue";
import { More, Odometer, Files, Collection } from '@element-plus/icons-vue';
import { useLayoutStore } from '@/stores/charts';
import { useTopoStore } from '@/stores/topo';
import MyEcharts from '@/views/MyEcharts.vue';
import { pickerOptions } from '@/utils/datePicker';
import { useMacStore } from "@/stores/mac";

let tags = reactive<any>([])
let table_data = reactive<any>([])
const display_drawer = ref(false);
const isHost = ref(false); // 节点类型是否为host
const chart = ref([] as any);
const icons = [More, Odometer, Files, Collection]
// node 基本信息
const node = reactive({
  type: '',
  name: '',
})
// 内层抽屉
const inner_drawer = reactive({
  display: false,
  type: '',
  label: ''
})

// 按钮组
const btns = reactive([
  {
    label: '指标详情',
    type: 'metric',
    disabled: false
  },
  {
    label: '图表选择',
    type: 'chart',
    disabled: false
  },
  {
    label: '配置文件',
    type: 'config',
    disabled: false
  },
  {
    label: '保存配置',
    type: 'save',
    disabled: false
  },
])

const layoutStore = useLayoutStore();
let layout = reactive(layoutStore.layout_option);

const grid = reactive({
  draggable: true,
  resizable: true,
  responsive: true,
});

// 监听topo数据
watch(() => useTopoStore().nodeData, (new_node_data, old_node_data) => {
  if (new_node_data.model) {
    let node_data = new_node_data.model;
    display_drawer.value = true;
    // 是否是host类型
    if (node_data.Type === 'host') {
      isHost.value = true;
      useMacStore().setMacIp(node_data.id.split("_")[2]);
    }
    node.name = node_data.label;
    node.type = node_data.Type;
    tags = node_data.tags;
    table_data = [];
    let metrics = node_data.metrics;
    for (let key in metrics) {
      table_data.push({
        name: key,
        value: metrics[key],
      })
    };
  }
}, {
  immediate: true
})

// 展示内层抽屉
const handleBtnClick = (btnItem: { type: string, label: string }) => {
  inner_drawer.display = btnItem.type === 'save' ? false : true;
  inner_drawer.type = btnItem.type;
  inner_drawer.label = btnItem.label;
}
const handleIcon = (index: number) => {
  return icons[index]
}

function handleClose() {
  display_drawer.value = false;
}

// 选择展示时间范围
let dateRange = ref([new Date() as any - 2 * 60 * 60 * 1000, new Date() as any - 0])
const startTime = ref(0);
const endTime = ref(0);
startTime.value = (new Date() as any) / 1000 - 60 * 60 * 2;
endTime.value = (new Date() as any) / 1000;
const changeDate = (value: number[]) => {
  if (value) {
    startTime.value = (new Date(value[0]) as any) / 1000;
    endTime.value = (new Date(value[1]) as any) / 1000;
  } else {
    startTime.value = (new Date() as any) / 1000 - 60 * 60 * 2;
    endTime.value = (new Date() as any) / 1000;
  }
}

// echarts大小随grid改变
const SizeAutoChange = (i: string, isChart?: boolean) => {
  if (isChart) {
    chart.value[i].resize();
  }
}
</script>

<style scoped>
.drawer {
  overflow: hidden;
}

.drawer_body_div {
  width: 100%;
  display: relative;
}

.drawer_top_div {
  display: flex;
  justify-content: space-between;
}

.drag {
  --title_height: 24px;
  width: 100%;
  height: var(--title_height);
  border-radius: 4px 4px 0 0;
  position: absolute;
  z-index: 9999;
  display: flex;
  align-items: center;
  justify-content: center;
}

.drag-title {
  display: flex;
  align-items: center;
  justify-content: center;
  user-select: none;
  width: 88%;
  height: 100%;
  color: #303133;
  font-size: 12px;
  font-weight: bold;
}

.drag:hover {
  background: rgba(253,
      186,
      74, .6)
}

.noDrag {
  --title_height: 24px;
  width: 100%;
  height: calc(100% - var(--title_height));
  margin-top: var(--title_height);
  display: flex;
  justify-content: center;
  align-items: center;
}

.noDrag-text {
  font-weight: bold;
  font-size: 20px;
  color: #67e0e3;
  user-select: none;
}

.vue-grid-layout {
  width: 100%;
  height: 100%;
  margin-top: 5px;
  background: #f1ecec;
}

.vue-grid-item {
  box-sizing: border-box;
  background-color: #ffffff;
  border-radius: 4px;
  box-shadow: 0 1px 5px rgba(45, 47, 51, 0.1);
}

.vue-grid-item .resizing {
  opacity: 0.9;
}

.vue-grid-item .static {
  background: #cce;
}

.vue-grid-item .text {
  font-size: 24px;
  text-align: center;
  position: absolute;
  top: 0;
  bottom: 0;
  left: 0;
  right: 0;
  margin: auto;
  height: 100%;
  width: 100%;
}

.vue-grid-item .no-drag {
  height: 100%;
  width: 100%;
}

.vue-grid-item .minMax {
  font-size: 12px;
}

.vue-grid-item .add {
  cursor: pointer;
}

.vue-draggable-handle {
  position: absolute;
  width: 20px;
  height: 20px;
  top: 0;
  left: 0;
  /* background: url("data:image/svg+xml;utf8,<svg xmlns='http://www.w3.org/2000/svg' width='10' height='10'><circle cx='5' cy='5' r='5' fill='#999999'/></svg>") no-repeat; */
  background-color: aqua;
  background-position: bottom right;
  padding: 0 8px 8px 0;
  background-repeat: no-repeat;
  background-origin: content-box;
  box-sizing: border-box;
  cursor: pointer;
}
</style>