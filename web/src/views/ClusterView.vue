<template>
  <div class="expand-collapse-button-container">
     <el-button class="expand-collapse-button" round size="large" @click="globalExpandAndCollapse">{{ global_combos }}</el-button> 
  </div>

  <div id="topo-container" class="container"></div>

  <HostDrawer :host_drawer="drawer_display['host']" :node="node" @update-statu="closeDrawer('host')"/>
  <ProcessDrawer :process_drawer="drawer_display['process']" :node="node" @update-state="closeDrawer('process')"/>
  
</template>

<script setup lang="ts">
import G6, { Graph } from '@antv/g6';
import { ref, reactive, onMounted, watch } from "vue";
import { topo } from '../request/api';
import server_logo from "@/assets/icon/server.png";
import topodata from '@/assets/cluster-2.json'
import { useMacStore } from '@/stores/mac';
import HostDrawer from '@/views/HostDrawer.vue'
import ProcessDrawer from '@/views/ProcessDrawer.vue'

const props = defineProps({
  graph_mode: {
    type: String,
    default: 'default',
    requires: true,
  },
})

const global_combos = ref('collapse')
let graph: Graph
let node: any
let data: any

let drawer_display = reactive({
  "host": false,
  'process': false,
  'thread': false,
  'net': false,
  'resource': false
})

const subjectColors = [
  '#5F95FF', // blue
  '#61DDAA',
  '#65789B',
  '#F6BD16',
  '#7262FD',
  '#78D3F8',
  '#9661BC',
  '#F6903D',
  '#008685',
  '#F08BB4',
];
const backColor = '#fff';
const theme = 'default';
const disableColor = '#777';
const colorSets = G6.Util.getColorSetsBySubjectColors(
  subjectColors,
  backColor,
  theme,
  disableColor,
);

onMounted(async () => {
  try {
    // ttcode
    data = topodata
    // const data = await topo.multi_host_topo();


    for (let i = 0; i < data.data.edges.length; i++) {
      let edge: any = data.data.edges[i];
      if (edge.Type === "belong") {
        edge.style = {
          stroke: "red",
          lineWidth: 2,
        }
      } else {
        edge.style = {
          lineWidth: 3,
        }
      }
    };

    for (let i = 0; i < data.data.nodes.length; i++) {
      let node: any = data.data.nodes[i];
      node.nodeStrength = -30;
      if (node.Type === "host") {
        node.img = server_logo;
        node.type = "image";
        node.size = 40;
        node.nodeStrength = -200;
        let ip = node.id.split("_").pop()
        node.label = ip;
      } else if (node.Type === "process") {
        node.label = node.name + ":" + node.metrics.Pid;
      } else if (node.Type === "net") {
        node.label = node.name;
      }
    };

    data.data.combos.forEach((combo: any, i: any) => {
    const color = colorSets[i % colorSets.length];
    combo.style = {
      stroke: color.mainStroke,
      fill: color.mainFill,
      opacity: 0.8
    }
    })

    initGraph(data.data);
    graph.setMode(props.graph_mode);

    data.data.combos.forEach((combo: any, i: any) => {
      graph.collapseCombo(combo['id']);
    })

  } catch (error) {
    console.error(error)
  }
})

function initGraph(data: any) {
  graph = new G6.Graph({
    container: "topo-container",
    width: document.getElementById("topo-container")!.clientWidth,
    height: document.getElementById("topo-container")!.clientHeight,
    fitView: true,
    fitViewPadding: 50,
    animate: true,
    minZoom: 0.00000001,
    layout: {
      pipes: [
        {
          type: 'comboCombined',
          outerLayout: new G6.Layout['gForce']({
            gravity: 1,
            factor: 2,
            preventOverlap: true,
            linkDistance: (edge: any, source: any, target: any) => {
              const nodeSize = ((source.size?.[0] || 30) + (target.size?.[0] || 30)) / 2;
              return Math.min(nodeSize * 1.5, 700);
            }
          }),
        },
        {
          type: 'radial',
          // center: [ 0, 0 ],
          focusNode: '54bcecd3-ea5f-497e-9ccb-3bb1aa9c0864_host_10.10.10.20',
          unitRadius: 150,
          maxIteration: 300,
          linkDistance: 30,
          preventOverlap: true,
          nodeSize: 30,
          sortBy: 'layoutattr',
          sortStrength: 50,
          nodesFilter: (node: any) => node.uuid === '54bcecd3-ea5f-497e-9ccb-3bb1aa9c0864',
        },
        {
          type: 'radial',
          // center: [ -800, 600],
          focusNode: '070cb0b4-c415-4b6a-843b-efc51cff6b76_host_10.10.10.60',
          unitRadius: 150,
          maxIteration: 300,
          linkDistance: 100,
          preventOverlap: true,
          nodeSize: 30,
          sortBy: 'layoutattr',
          sortStrength: 50,
          nodesFilter: (node: any) => node.uuid === '070cb0b4-c415-4b6a-843b-efc51cff6b76',
        },
        {
          type: 'radial',
          // center: [700, 500],
          focusNode: '7d0740a7-5ee6-41a9-846b-d52890d690d5_host_10.10.10.111',
          unitRadius: 150,
          maxIteration: 300,
          linkDistance: 100,
          preventOverlap: true,
          nodeSize: 30,
          sortBy: 'layoutattr',
          sortStrength: 50,
          nodesFilter: (node: any) => node.uuid === '7d0740a7-5ee6-41a9-846b-d52890d690d5',
        },
      ],
    },
    modes: {
      default: ['drag-canvas', 'drag-combo', 'zoom-canvas', 'collapse-expand-combo'],
      localmode: ['drag-canvas', 'zoom-canvas', "drag-node", 'lasso-select', 'brush-select', "click-select"],
      mixmode: ['drag-canvas', 'zoom-canvas', 'drag-combo', 'collapse-expand-combo', "drag-node", "click-select"]
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
      node = (e.target as any)._cfg
      console.log("click node:", node.id);
      switch (node.model.Type) {
        case 'host':
          useMacStore().setMacIp(node.id.split("_")[2]);
          drawer_display['host'] = true
          break;
        case 'process':
          drawer_display['process'] = true
          break;
      }

    } else {
      console.log("node unselected")
    }

    return false
  });
  graph.on('node:dragstart', (e) => {
    // graph.layout();
    refreshDragedNodePosition(e);
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

function refreshDragedNodePosition(e: any) {
  const model = e.item.get('model');
  model.fx = e.x;
  model.fy = e.y;
}

function closeDrawer(nodetype: string) {
  switch (nodetype) {
    case 'host':
      drawer_display['host'] = false
      graph.findAllByState('node', 'selected').forEach(( item: any, i: any ) => {
        graph.setItemState(item, 'selected', false)
      })
      break;
    case 'process':
      drawer_display['process'] = false
      graph.findAllByState('node', 'selected').forEach(( item: any, i: any ) => {
        graph.setItemState(item, 'selected', false)
      })
      break;
    case 'thread':
      drawer_display['thread'] = false
      break;
    case 'net':
      drawer_display['net'] = false
      break;
    case 'resource':
      drawer_display['resource'] = false
      break;
  }
}

function globalExpandAndCollapse() {
  global_combos.value = global_combos.value === 'collapse' ? 'expand' : 'collapse'
  
  if (global_combos.value === 'collapse' ) {
    data.data.combos.forEach((combo: any, i: any) => {
      graph.collapseCombo(combo['id']);
    })
    graph.updateCombos();
  } else {
    data.data.combos.forEach((combo: any, i: any) => {
      graph.expandCombo(combo['id']);
    })
    graph.updateCombos();
  }
}

watch(() => props.graph_mode, (new_data) => {
  graph.setMode(new_data)
}, {
  deep: true
})

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
    padding: 10px;
  }
  
.drawer_head_div {
    width: 100%;
    height: 200px;
    margin-top: 35px;

    display: absolute;
    border-bottom: 1px solid rgb(181, 177, 177);

    /* border: 1px groove rgb(195, 184, 184);
    border-radius: 10px; */
  }

.tag {
  font-size: 14px;
  font-weight: bold;
  border-spacing: 5px;
  border-radius: 8px;
  padding: 4px;

  color: #ffffff;
}

.drawer_body_div {
  width: 100%;
  /* height: 80%; */

  display: relative;
}

.nested_metric_drawer_div {
  position: absolute;
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

.expand-collapse-button-container {
  display: flex;
  justify-content: center;
  /* align-items: center; */
}

.expand-collapse-button {
  position: absolute;
  bottom: 0;
  margin-bottom: 100px;
  background-color: #67e0e3;
}

</style>
