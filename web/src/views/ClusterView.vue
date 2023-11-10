<template>
  <div class="title">
    <h1 class="h1">集群拓扑图演示页面</h1>
    <el-button class="button" @click="switch_single_node">单机拓扑</el-button>
  </div>
  <div id="topo-container" class="container"></div>
  <el-drawer class="drawer" v-model="drawer" :title="title" direction="rtl" size="40%">
    <div>
      <el-button @click="innerDrawer = true">
        <el-icon class="el-icon--left"><More /></el-icon>
      </el-button>
      <el-drawer
        v-model="innerDrawer"
        title="I'm inner Drawer"
        :append-to-body="true"
         >
        <el-table :data="table_data" stripe style="width: 100%">
          <el-table-column prop="name" label="属性" width="180" />
          <el-table-column prop="value" label="值" />
        </el-table>
      </el-drawer>
    </div>
  </el-drawer>
</template>

<script setup lang="ts">
import G6 from '@antv/g6';
import { ref, reactive, onMounted } from "vue";
import { useRouter } from "vue-router";
import { topo } from '../request/api';
import server_logo from "@/assets/icon/server.png";
import type { More } from '@element-plus/icons-vue';

let drawer = ref(false)
const innerDrawer = ref(false)
let title = ref("")
let table_data = reactive<any>([])

const router = useRouter()

function handleClose() {
  drawer.value = false
}

function switch_single_node() {
  router.push("/node")
}

onMounted(async () => {
  try {
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
      type: 'force',
      preventOverlap: true,
      linkDistance: 100,
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
  title.value = node.id + "节点属性";
  drawer.value = drawer.value ? false : true;

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

</script>

<style scoped>
.title {
  position: relative;
}

.h1 {
  width: 100%;
  text-align: center;
}

.button {
  font-size: 120%;
  position: absolute;
  background-color: white;
  right: 0;
  bottom: 0;

  margin-bottom: 3px;
  margin-right: 10px;
  padding-left: 10px;
}

.container {
  width: 100%;
  height: 100%;
  background-color: white;
}

.drawer {
  height: 100%;
}
</style>
