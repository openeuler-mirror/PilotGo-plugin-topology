<template>
  <div id="topo-container" class="container"></div>
  <el-drawer class="drawer" v-model="chart_drawer" :with-header="false" direction="rtl" size="30%">
    <div class="drawer_head"></div>
    <div>
      <el-button class="metric_button" @click="metric_drawer_inner = true" :icon="More" size="small" circle="true"/>
        <!-- <el-icon class="el-icon--left"><More /></el-icon> -->
      <el-drawer v-model="metric_drawer_inner" :with-header="false" :append-to-body="true" size="25%">
        <el-table :data="table_data" stripe style="width: 100%">
          <el-table-column prop="name" label="属性" />
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
import { More, Menu } from '@element-plus/icons-vue';
import topodata from '../../public/cluster.json'

let chart_drawer = ref(false)
let metric_drawer_inner = ref(false)
let title = ref("")
let table_data = reactive<any>([])

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
      edgeStrength: 50,
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
  title.value = node.id;

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

</script>

<style scoped>

.container {
  width: 100%;
  height: 100%;
  background-color: white;
}

.drawer {
    height: 100%;
  }

.drawer_head {
      width: 100%;
      height: 20%;

      border-bottom: 1px solid rgb(181, 177, 177);
    }

.metric_button {
      position: absolute;
      right: 0;
      top: 0; 

      margin-top: 170px;
      margin-right: 10px;
    }


</style>
