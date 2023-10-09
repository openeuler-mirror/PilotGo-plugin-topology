<template>
  <div class="title">
    <h1 class="h1">单机拓扑图演示页面</h1>
    <el-dropdown class="dropdown" @command="handleNodeSelected">
      <span class="el-dropdown-link">
        选择主机
        <el-icon class="el-icon--right">
          <arrow-down />
        </el-icon>
      </span>
      <template #dropdown>
        <el-dropdown-menu>
          <el-dropdown-item :command=node v-for="node in node_list">{{ node.id }}</el-dropdown-item>
        </el-dropdown-menu>
      </template>
    </el-dropdown>
    <el-button class="button" @click="switch_cluster">集群拓扑</el-button>
  </div>

  <div id="topo-container" class="container"></div>
  <el-drawer class="drawer" v-model="drawer" :title="title" direction="rtl" :before-close="handleClose">
    <el-table :data="table_data" stripe style="width: 100%">
      <el-table-column prop="name" label="属性" width="180" />
      <el-table-column prop="value" label="值" />
    </el-table>
  </el-drawer>
</template>

<script setup lang="ts">
import G6 from '@antv/g6';
import { ref, reactive, onMounted } from "vue";
import { useRouter } from "vue-router";
import { topo } from '../request/api';
import server_logo from "@/assets/icon/server.png";

const router = useRouter()

const node_list = reactive<any>([])
let graph = ref()

let drawer = ref(false)
let title = ref("")
let table_data = reactive<any>([])

function handleClose() {
  drawer.value = false
}

function switch_cluster() {
  router.push("/cluster")
}

onMounted(async () => {
  try {
    updateNodeList()
  } catch (error) {
    console.error(error)
  }
})

async function updateNodeList() {
  const data = await topo.host_list()
  // console.log(data);
  for (let key in data.data.agentlist) {
    node_list.push({
      id: key,
    })
  };

}

async function handleNodeSelected(node: any) {
  const data = await topo.single_host_tree(node.id);
  // console.log(data.data.tree);

  let root = data.data.tree
  root.img = server_logo;
  root.type = "image";
  root.size = 40

  initGraph(data.data.tree);
}

function initGraph(data: any) {
  if (graph.value != null) {
    graph.value.destroy();
  }
  graph.value = new G6.TreeGraph({
    container: "topo-container",
    width: document.getElementById("topo-container")!.clientWidth,
    height: document.getElementById("topo-container")!.clientHeight,
    modes: {
      default: ['drag-canvas', 'zoom-canvas', "click-select", "drag-node",
        {
          type: 'collapse-expand',
          onChange: function onChange(item: any, collapsed) {
            const data = item.getModel();
            data.collapsed = collapsed;
            return true;
          },
        },
      ],
    },
    layout: {
      type: 'dendrogram',
      direction: 'LR',
      nodeSep: 30,
      rankSep: 100,
    },
  });
  graph.value.node(function (node: any) {
    // console.log(node);
    return {
      label: node.node.type + ":" + node.node.name,
      labelCfg: {
        position: node.children && node.children.length > 0 ? 'left' : 'right',
        offset: 5,
      },
    };
  });
  graph.value.on("nodeselectchange", (e: any) => {
    if (e.select) {
      let node = (e.target as any)._cfg
      console.log("click node:", node.id);

      updateDrawer(node)
    } else {
      console.log("node unselected")
    }
    return false
  });
  graph.value.data(data);
  graph.value.render();
  graph.value.fitCenter();

  window.onresize = () => {
    graph.value.changeSize(
      document.getElementById("topo-container")!.clientWidth,
      document.getElementById("topo-container")!.clientHeight)
    graph.value.fitCenter()
  }
}

function updateDrawer(node: any) {
  title.value = node.id + "节点属性";
  drawer.value = drawer.value ? false : true;

  // console.log(node);
  table_data = [];
  let metrics = node.model.node.metrics;
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
  margin: 0;
  padding-top: 10px;
  padding-bottom: 10px;
  text-align: center;
}

.dropdown {
  font-size: 120%;
  position: absolute;
  background-color: white;
  right: 120px;
  bottom: 5px;

  margin-bottom: 3px;
  margin-right: 10px;
  padding-left: 10px;
}

.button {
  font-size: 120%;
  position: absolute;
  background-color: white;
  right: 0;
  bottom: 0;

  margin-bottom: 0px;
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
