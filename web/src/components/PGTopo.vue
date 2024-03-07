<template>
  <div v-loading="loading" id="topo-container" class="container"></div>
</template>

<script setup lang="ts">
import G6, { Graph } from '@antv/g6';
import { ref, reactive, onMounted, watch, watchEffect } from "vue";
import server_logo from "@/assets/icon/server_blue.png";
import process_logo from "@/assets/icon/process.png";
import net_logo from "@/assets/icon/net.png";
import resource_logo from "@/assets/icon/resource.png";
import { useTopoStore } from '@/stores/topo';
import { colorSets, graphInitOptions, graphTreeInitOptions } from './PGOptions';

const props = defineProps({
  graph_mode: {
    type: String,
    default: 'default',
    require: true
  },
  time_interval: {
    type: String,
    default: '关闭',
    require: true
  }
})

let graph: Graph;
const init_data = ref(false);
let topo_data = reactive<any>({});
let topo_type = ref('comb');
const loading = ref(false);

const topoW = ref(0);
const topoH = ref(0);

onMounted(() => {
  loading.value = true;
  topoW.value = document.getElementById("topo-container")!.clientWidth;
  topoH.value = document.getElementById("topo-container")!.clientHeight;
})

const updateTopoData = (topoData: any) => {
  topo_data = topoData;

  topo_data.edges.forEach((item: any) => {
    item.style = { lineWidth: 3 };
    if (item.Type === "belong") {
      item.style = {
        stroke: "red",
        lineWidth: 1,
      }
    }
  })

  topo_data.nodes.forEach((item: any) => {
    item.nodeStrength = -30;
    switch (item.Type) {
      case "host":
        item.img = server_logo;
        item.type = "image";
        item.size = 40;
        item.nodeStrength = -200;
        item.label = item.id.split("_").pop();
        break;
      case "process":
        item.img = process_logo;
        item.type = "image";
        item.label = item.name + ":" + item.metrics.Pid;
        item.style = {
          stroke: '#7262FD'
        };
        item.labelCfg = {
          position: "right",
          offset: 2,
        };
        break;
      case "net":
        item.img = net_logo;
        item.type = "image";
        item.label = item.name;
        item.style = {
          stroke: '#F6BD16'
        };
        break;
      default:
        // resource
        item.img = resource_logo;
        item.type = "image";
        item.label = item.name;
        item.style = {
          stroke: '#78D3F8'
        };
        item.labelCfg = {
          position: "left",
          offset: 2,
        };
        break;
    }
  })

  topo_data.combos.forEach((combo: any, i: number) => {
    combo.style = {
      stroke: colorSets[i].mainStroke,
      fill: colorSets[i].mainFill,
      opacity: 0.8
    }
  })

  // 数据处理完成，初始化图
  init_data.value = true;
}

function initGraph(data: any) {
  if (graph) {
    graph.destroy();
    // graph.refresh();
  }
  let graphBox = {
    container: "topo-container",
    width: topoW.value,
    height: topoH.value,
  }

  if (topo_type.value === 'comb') {
    graph = new G6.Graph({
      ...graphBox,
      ...graphInitOptions,
    });
  } else {
    graph = new G6.TreeGraph({
      ...graphBox,
      ...graphTreeInitOptions,
    });
    graph.node(function (node: any) {
      return {
        label: node.node.Type + ":" + node.node.name,
        labelCfg: {
          position: node.children && node.children.length > 0 ? 'left' : 'right',
          offset: 5,
        },
      };
    });
  }
  // 节点点击事件
  graph.on('node:click', (e: any) => {
    graph.getNodes().forEach((node) => {
      graph.clearItemStates(node);
    });
    const nodeItem = e.item;
    graph.setItemState(nodeItem, 'click', true);
    // 抽屉组件展示的节点数据
    let selected_node = e.item._cfg;
    if (topo_type.value === 'comb') {
      useTopoStore().nodeData = selected_node.model;
    } else {
      useTopoStore().nodeData = selected_node.model.node;
    }
  });
  // 节点悬浮高亮
  graph.on('node:mouseover', (e: any) => {
    graph.setItemState(e.item, 'active', true);
  });
  // 节点鼠标移出后取消节点悬浮高亮
  graph.on('node:mouseout', (e: any) => {
    graph.setItemState(e.item, 'active', false);
  });

  graph.on('node:dragstart', (e) => {
    // graph.layout();
    refreshDragedNodePosition(e);
  });
  // 解决拖动产生残影问题
  graph.get('canvas').set('localRefresh', false);

  graph.data(data);

  graph.render();



  window.onresize = () => {
    graph.changeSize(
      document.getElementById("topo-container")!.clientWidth,
      document.getElementById("topo-container")!.clientHeight)
    graph.fitCenter()
  }
  loading.value = false;
  init_data.value = false
}

function refreshDragedNodePosition(e: any) {
  const model = e.item.get('model');
  model.fx = e.x;
  model.fy = e.y;
}

watchEffect(() => {
  // 数据处理的入口
  topo_type.value = useTopoStore().topo_type;
  let topo_data = JSON.parse(JSON.stringify(useTopoStore().topo_data));
  if (topo_data.tree || topo_data.nodes) {
    if (topo_type.value === 'tree') {
      initGraph(topo_data.tree);
    } else {
      updateTopoData(topo_data);
    }
  }
})

watch(() => init_data, (newdata) => {
  if (newdata) {
    initGraph(topo_data);
    topo_data.combos.forEach((combo: any, i: any) => {
      graph.collapseCombo(combo['id']);
    })
    graph.updateCombos();
  }

}, { deep: true })


// 设置graph_mode
watch(() => props.graph_mode, (newData) => {
  if (newData) {
    graph.setMode(newData);
  }
})

</script>

<style scoped>
.container {
  width: 100%;
  height: 100%;
  background-color: white;
}
</style>
