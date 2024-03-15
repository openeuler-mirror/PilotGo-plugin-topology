<template>
  <div id="topo-container" class="container"></div>
</template>

<script setup lang="ts">
import G6, { Graph } from '@antv/g6';
import { ref, reactive, onMounted, watch, watchEffect } from "vue";
import server_logo from "@/assets/icon/server_blue.png";
import process_logo from "@/assets/icon/process.png";
import net_logo from "@/assets/icon/net.png";
import resource_logo from "@/assets/icon/resource.png";
import machine_logo from "@/assets/icon/machine.png";
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

const topoW = ref(0);
const topoH = ref(0);

onMounted(() => {
  init_data.value = false;
  topoW.value = document.getElementById("topo-container")!.clientWidth;
  topoH.value = document.getElementById("topo-container")!.clientHeight;
})

const updateTopoData = (topoData: any) => {
  topo_data = topoData;

  topo_data.edges.forEach((item: any) => {
    item.style = {
       lineWidth: 3,
       cursor: "all-scroll",
    };
    item.labelCfg = {
      position: 'middle',
      // offset: 2,
      autoRotate: true,
      style: {
        fontSize: 12,
        opacity: 0.5,
      },
    
    };
    switch (item.Type) {
      case "belong":
        item.style = {
          stroke: "red",
          lineWidth: 1,
          cursor: "all-scroll",
        };
        // item.label = item.Type;
        break;
      case "tcp":
        // item.label = item.id.split("__")[1];
        item.label = item.Type;
        break;
      case "udp":
        // item.label = item.id.split("__")[1];
        item.label = item.Type;
        break;
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
        item.style = {
          cursor: "all-scroll"
        };
        break;
      case "process":
        item.img = process_logo;
        item.type = "image";
        item.label = item.name + ":" + item.metrics.Pid;
        item.style = {
          stroke: '#7262FD',
          cursor: "all-scroll"
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
          stroke: '#F6BD16',
          cursor: "all-scroll"
        };
        break;
      default:
        // resource
        item.img = resource_logo;
        item.type = "image";
        item.label = item.name;
        item.style = {
          stroke: '#78D3F8',
          cursor: "all-scroll"
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
    combo.collapsedSubstituteIcon = {
        show: true,
        img: machine_logo,
        width: 50,
        height: 50
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
  // 边点击事件
  graph.on('edge:click', (e: any) => {
    graph.getEdges().forEach((edge) => {
      graph.clearItemStates(edge);
    });
    const edgeItem = e.item;
    graph.setItemState(edgeItem, 'click', true);
    // 抽屉组件展示的边数据
    let selected_edge = e.item._cfg;
    if (topo_type.value === 'comb') {
      useTopoStore().edgeData = selected_edge.model;
    } else {
      useTopoStore().edgeData = selected_edge.model.edge;
    }
  });
  // // 节点悬浮高亮
  // graph.on('node:mouseover', (e: any) => {
  //   graph.setItemState(e.item, 'active', true);
  // });
  // // 节点鼠标移出后取消节点悬浮高亮
  // graph.on('node:mouseout', (e: any) => {
  //   graph.setItemState(e.item, 'active', false);
  // });

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
  init_data.value = false
}

function refreshDragedNodePosition(e: any) {
  const model = e.item.get('model');
  model.fx = e.x;
  model.fy = e.y;
}

watch(() => useTopoStore().topo_data, (new_topo_data) => {
  // 数据处理入口
  topo_type.value = 'comb';
  let topo_data = JSON.parse(JSON.stringify(new_topo_data));
  if (topo_data.tree) {
    topo_type.value = 'tree';
    initGraph(topo_data.tree);
  } else if (topo_data.nodes) {
    updateTopoData(topo_data);
  }
}, { immediate: true, deep: true })


watch(() => init_data, (newdata) => {
  if (newdata) {
    initGraph(topo_data);
    topo_data.combos.forEach((combo: any, _i: any) => {
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
