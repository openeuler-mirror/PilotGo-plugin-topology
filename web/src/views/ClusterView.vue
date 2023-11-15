<template>
  <div id="topo-container" class="container"></div>
  <el-drawer class="drawer" v-model="chart_drawer" :with-header="false" direction="rtl" size="30%">
    <div class="drawer_head_div"></div>

    <div class="drawer_body_div">
      <div :style="{ 'position': 'relative', 'display': 'flex', 'flex-direction': 'column', 'align-items': 'center' }">
        <div class="drag">
          <span class="drag-title">{{ layout[0].title }}</span>
        </div>
        <div class="noDrag">
          <MyEcharts :query="layout[0].query" :startTime="startTime"
            :endTime="endTime" style="width:100%;height:100%;">
          </MyEcharts>
          <!-- <TempEcharts style="width:100%;height:100%;">
          </TempEcharts> -->
        </div>
      </div>
    </div>

    <div class="drawer_inner_div">
      <el-drawer v-model="metric_drawer_inner" :with-header="false" :append-to-body="true" size="25%">
        <el-table :data="table_data" stripe style="width: 100%">
          <el-table-column prop="name" label="属性" />
          <el-table-column prop="value" label="值" />
        </el-table>
      </el-drawer>
    </div>

    <div class="drawer_top_div" :style="{ 'margin-top': '10px', 'margin-right': '10px' }">
      <!-- 时间范围选择 -->
      <el-date-picker  v-model="dateRange" type="datetimerange" :shortcuts="pickerOptions" range-separator="to"
        start-placeholder="开始日期" end-placeholder="结束日期" @change="changeDate" size="small"
        :style="{ 'width': '290px', 'height': '34px', 'margin-right': '10px' }">
      </el-date-picker>
      <!-- 指标数据 -->
      <el-button class="drawer_button" @click="metric_drawer_inner = true" :icon="More" size="default" circle="true" />
      <!-- 选择要显示的图表 -->
      <el-button class="drawer_button" @click="chart_drawer_inner = true" :icon="Platform" size="default" circle="true" />
      <!-- 加载本地的图表配置文件 -->
      <el-button class="drawer_button" @click="chart_drawer_inner = true" :icon="Files" size="default" circle="true" />
    </div>

  </el-drawer>
</template>

<script setup lang="ts">
import G6 from '@antv/g6';
import { ref, reactive, onMounted } from "vue";
import { useRouter } from "vue-router";
import { topo } from '../request/api';
import server_logo from "@/assets/icon/server.png";
import { More, Platform, Files } from '@element-plus/icons-vue';
import topodata from '@/assets/cluster.json'
import { useLayoutStore } from '@/stores/charts';
import MyEcharts from '@/views/MyEcharts.vue';
import { pickerOptions } from '@/utils/datePicker';
import TempEcharts from './TempEcharts.vue'

let chart_drawer = ref(false)
let metric_drawer_inner = ref(false)
let chart_drawer_inner = ref(false)
let table_data = reactive<any>([])
let dateRange = ref([new Date() as any - 2 * 60 * 60 * 1000, new Date() as any - 0])

const startTime = ref(0);
const endTime = ref(0);
startTime.value = (new Date() as any) / 1000 - 60 * 60 * 2;
endTime.value = (new Date() as any) / 1000;

const layoutStore = useLayoutStore();
let layout = reactive(layoutStore.layout_option);

const router = useRouter()

function handleClose() {
  chart_drawer.value = false
}

onMounted(async () => {
  try {
    // ttcode
    // const data = topodata

    const data = await topo.multi_host_topo();
    // console.log(data.data);

    for (let i = 0; i < data.data.edges.length; i++) {
      let edge: any = data.data.edges[i];
      if (edge.Type === "belong") {
        edge.style = {
          stroke: "red",
          lineWidth: 2,
        }
      } else if (edge.Type === "server") {

      }
    };

    for (let i = 0; i < data.data.nodes.length; i++) {
      let node: any = data.data.nodes[i];
      node.nodeStrength = -30;
      if (node.type === "host") {
        node.img = server_logo;
        node.type = "image";
        node.size = 40;
        node.nodeStrength = -200;
        let ip = node.id.split("_").pop()
        node.label = ip;
      } else if (node.type === "process") {
        node.label = node.name + ":" + node.metrics.Pid;
      } else if (node.type === "net") {
        node.label = node.name;
      }
    };

    initGraph(data.data);
  } catch (error) {
    console.error(error)
  }
})

function initGraph(data: any) {
  let graph = new G6.Graph({
    container: "topo-container",
    width: document.getElementById("topo-container")!.clientWidth,
    height: document.getElementById("topo-container")!.clientHeight,
    layout: {
      // type: 'force',
      // preventOverlap: true,
      // linkDistance: 100,

      type: 'gForce',
      gravity: 0.1,
      edgeStrength: 10,
      nodeStrength: 100,

    },
    modes: {
      default: ['drag-canvas', 'zoom-canvas', "click-select", "drag-node"],
    },
  });
  graph.node(function (node) {
    return {
      labelCfg: {
        position: "bottom",
        offset: 5,
      },
    };
  });
  graph.on("nodeselectchange", (e) => {
    if (e.select) {
      let node = (e.target as any)._cfg
      console.log("click node:", node.id);

      updateDrawer(node)
    } else {
      console.log("node unselected")
    }
    return false
  });
  graph.on('node:dragstart', (e) => {
    graph.layout();
  });
  graph.data(data);
  graph.render();

  window.onresize = () => {
    graph.changeSize(
      document.getElementById("topo-container")!.clientWidth,
      document.getElementById("topo-container")!.clientHeight)
    graph.fitCenter()
  }
}

function updateDrawer(node: any) {
  // if (node.type === "host") {
  //     chart_drawer.value = chart_drawer.value ? false : true;
  // } else {
  //     metric_drawer.value = metric_drawer.value ? false : true;
  // }

  chart_drawer.value = chart_drawer.value ? false : true;

  // console.log(node)
  table_data = [];
  let metrics = node.model.metrics;
  for (let key in metrics) {
    table_data.push({
      name: key,
      value: metrics[key],
    })
  };
}

// 选择展示时间范围
const changeDate = (value: number[]) => {
  if (value) {
    startTime.value = (new Date(value[0]) as any) / 1000;
    endTime.value = (new Date(value[1]) as any) / 1000;
  } else {
    startTime.value = (new Date() as any) / 1000 - 60 * 60 * 2;
    endTime.value = (new Date() as any) / 1000;
  }
}

</script>

<style scoped>

.container {
  width: 100%;
  height: 100%;
  background-color: white;
}

.drawer {
    position: relative;
    height: 100%;
  }

.drawer_head_div {
    width: 100%;
    height: 20%;

    display: absolute;
    border-bottom: 1px solid rgb(181, 177, 177);
  }

.drawer_body_div {
  width: 80%;
  height: 20%;

  display: absolute;
}

.drawer_inner_div {
  position: relative;
}

.drawer_top_div {
  position: absolute;
  right: 0;
  top: 0; 
  display: flex;
  justify-content: space-between;
}

.drawer_button {
  background-color: #cfcaca;
}

.drag {
  width: 100%;
  height: var(--title_height);
  border-radius: 4px 4px 0 0;
  position: absolute;
  z-index: 9999;
  display: flex;
  align-items: center;
  justify-content: center;

  &-title {
    display: flex;
    align-items: center;
    justify-content: center;
    user-select: none;
    width: 88%;
    height: 100%;
    color: #303133;
    font-size: 12px;
    font-weight: bold;

    &:hover {
      cursor: move;
    }
  }

  &-more {
    width: 12%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    user-select: none;
    cursor: pointer;
  }

  &:hover {
    background: rgba(253,
        186,
        74, .6)
  }
}  

.noDrag {
  width: 100%;
  height: calc(100% - var(--title_height));
  margin-top: var(--title_height);
  display: flex;
  justify-content: center;
  align-items: center;
  &-text {
    font-weight: bold;
    font-size: 20px;
    color: #67e0e3;
    user-select: none;
  }
} 
</style>
