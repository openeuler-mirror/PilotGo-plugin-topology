<template>
  <div class="expand-collapse-button-container">
     <el-button class="expand-collapse-button" round size="large" @click="globalExpandAndCollapse">{{ global_combos }}</el-button> 
  </div>

  <div id="topo-container" class="container"></div>

  <HostDrawer :host_drawer="drawer_display['host']" :node="selected_node.model" @update-statu="closeDrawer('host')"/>
  <ProcessDrawer :process_drawer="drawer_display['process']" :node="selected_node.model" @update-statu="closeDrawer('process')"/>
  <NetDrawer :net_drawer="drawer_display['net']" :node="selected_node.model" @update-statu="closeDrawer('net')"/>
  
</template>

<script setup lang="ts">
import G6, { Graph } from '@antv/g6';
import { ref, reactive, onMounted, watch } from "vue";
import { topo } from '../request/api';
import server_logo from "@/assets/icon/server.png";
// import topodata from '@/assets/cluster-2.json'
// import topodata from '@/assets/cluster-test.json'
import { useMacStore } from '@/stores/mac';
import HostDrawer from '@/views/HostDrawer.vue'
import ProcessDrawer from '@/views/ProcessDrawer.vue'
import NetDrawer from '@/views/NetDrawer.vue'
import { number } from 'echarts';

const topo_time_emit = defineEmits(['update-topo-time'])

const props = defineProps({
    graph_mode: {
      type: String,
      default: 'default',
      requires: true
    },
    time_interval: {
      type: String,
      default: '关闭',
      requires: true
    }
})

const global_combos = ref('collapse')
let graph: Graph
let selected_node = reactive<any>({})
let init_collapse = false
const init_data = ref(false)
let topo_data = reactive<any>({})

let time_interval_num: number = 0
let timer: number = 0
const topo_time = ref('')

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
        updateTopoData()
      
    } catch (error) {
        console.error(error)
    }
})

async function updateTopoData() {
    // ttcode
    // topo_data = topodata
    topo_data = await topo.multi_host_topo();
    const unix_time = new Date(parseInt(topo_data.data.nodes[0].unixtime) * 1000)
    topo_time.value = unix_time.getFullYear().toString() + "/" +
                      (unix_time.getMonth() + 1).toString() + "/" +
                      unix_time.getDate().toString() + "-" +
                      unix_time.getHours().toString() + ":" +
                      unix_time.getMinutes().toString() + ":" +
                      unix_time.getSeconds().toString();
    topo_time_emit('update-topo-time', topo_time.value)

    for (let i = 0; i < topo_data.data.edges.length; i++) {
      let edge: any = topo_data.data.edges[i];
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

    for (let i = 0; i < topo_data.data.nodes.length; i++) {
      let node: any = topo_data.data.nodes[i];
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

    topo_data.data.combos.forEach((combo: any, i: any) => {
    const color = colorSets[i % colorSets.length];
    combo.style = {
      stroke: color.mainStroke,
      fill: color.mainFill,
      opacity: 0.8
    }
    })

	init_data.value = true
}

function initGraph(data: any) {
    if (graph) {
        graph.destroy();
        // graph.refresh();
    }

    graph = new G6.Graph({
      container: "topo-container",
      width: document.getElementById("topo-container")!.clientWidth,
      height: document.getElementById("topo-container")!.clientHeight,
      fitView: true,
      fitViewPadding: 200,
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
            innerLayout: new G6.Layout['radial']({
              unitRadius: 150,
              maxIteration: 300,
              linkDistance: 30,
              preventOverlap: true,
              nodeSize: 30,
              sortBy: 'layoutattr',
              sortStrength: 50,
            }),
          },
          // {
          //   type: 'radial',
          //   // center: [ 0, 0 ],
          //   focusNode: '54bcecd3-ea5f-497e-9ccb-3bb1aa9c0864_host_10.10.10.20',
          //   unitRadius: 150,
          //   maxIteration: 300,
          //   linkDistance: 30,
          //   preventOverlap: true,
          //   nodeSize: 30,
          //   sortBy: 'layoutattr',
          //   sortStrength: 50,
          //   nodesFilter: (node: any) => node.uuid === '54bcecd3-ea5f-497e-9ccb-3bb1aa9c0864',
          // },
          // {
          //   type: 'radial',
          //   // center: [ 1100, 300 ],
          //   focusNode: '070cb0b4-c415-4b6a-843b-efc51cff6b76_host_10.41.107.200',
          //   unitRadius: 150,
          //   maxIteration: 300,
          //   linkDistance: 100,
          //   preventOverlap: true,
          //   nodeSize: 30,
          //   sortBy: 'layoutattr',
          //   sortStrength: 50,
          //   nodesFilter: (node: any) => node.uuid === '070cb0b4-c415-4b6a-843b-efc51cff6b76',
          // },
          // {
          //   type: 'radial',
          //   // center: [ -1200, 300 ],
          //   focusNode: '7d0740a7-5ee6-41a9-846b-d52890d690d5_host_10.10.10.111',
          //   unitRadius: 150,
          //   maxIteration: 300,
          //   linkDistance: 100,
          //   preventOverlap: true,
          //   nodeSize: 30,
          //   sortBy: 'layoutattr',
          //   sortStrength: 50,
          //   nodesFilter: (node: any) => node.uuid === '7d0740a7-5ee6-41a9-846b-d52890d690d5',
          // },

          // {
          //   type: 'radial',
          //   center: [ 0, 0 ],
          //   focusNode: '637c21cc-2194-11b2-a85c-e5a789c3f0aa_host_10.41.107.34',
          //   unitRadius: 150,
          //   maxIteration: 300,
          //   linkDistance: 30,
          //   preventOverlap: true,
          //   nodeSize: 30,
          //   sortBy: 'layoutattr',
          //   sortStrength: 50,
          //   nodesFilter: (node: any) => node.uuid === '637c21cc-2194-11b2-a85c-e5a789c3f0aa',
          // },
          // {
          //   type: 'radial',
          //   center: [ 1100, 300 ],
          //   focusNode: 'caf5f9ae-f6b9-4fd8-80f6-6b12ce69feb4_host_10.41.107.200',
          //   unitRadius: 150,
          //   maxIteration: 300,
          //   linkDistance: 100,
          //   preventOverlap: true,
          //   nodeSize: 30,
          //   sortBy: 'layoutattr',
          //   sortStrength: 50,
          //   nodesFilter: (node: any) => node.uuid === 'caf5f9ae-f6b9-4fd8-80f6-6b12ce69feb4',
          // },
          // {
          //   type: 'radial',
          //   center: [ -1200, 300 ],
          //   focusNode: '7fa8dcb0-4953-4da1-a514-6bf07303c4c9_host_10.41.107.201',
          //   unitRadius: 150,
          //   maxIteration: 300,
          //   linkDistance: 100,
          //   preventOverlap: true,
          //   nodeSize: 30,
          //   sortBy: 'layoutattr',
          //   sortStrength: 50,
          //   nodesFilter: (node: any) => node.uuid === '7fa8dcb0-4953-4da1-a514-6bf07303c4c9',
          // },

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
            selected_node = (e.target as any)._cfg
            console.log("click node:", selected_node);

            switch (selected_node.model.Type) {
                case 'host':
                    useMacStore().setMacIp(selected_node.id.split("_")[2]);
                    drawer_display['host'] = true
                    break;
                case 'process':
                    drawer_display['process'] = true
                    break;
                case 'net':
                    drawer_display['net'] = true
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

    console.log('渲染方法：', data.nodes[1].unixtime)
    graph.data(data);
    graph.clear();
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
      graph.findAllByState('node', 'selected').forEach(( item: any, i: any ) => {
        graph.setItemState(item, 'selected', false)
      })
      break;
    case 'resource':
      drawer_display['resource'] = false
      break;
  }
}

function globalExpandAndCollapse() {
    global_combos.value = global_combos.value === 'collapse' ? 'expand' : 'collapse'
    
    if (global_combos.value === 'collapse' ) {
      topo_data.data.combos.forEach((combo: any, i: any) => {
        graph.collapseCombo(combo['id']);
      })
      graph.updateCombos();
    } else {
      topo_data.data.combos.forEach((combo: any, i: any) => {
        graph.expandCombo(combo['id']);
      })
      graph.updateCombos();
    }
}

function timerhandler() {
    updateTopoData();
}

watch(() => props.graph_mode, (newdata) => {
    graph.setMode(newdata)
})

watch(() => init_data, (newdata) => {
	if (newdata) {
		initGraph(topo_data.data);
		graph.setMode(props.graph_mode);

		if (!init_collapse) {
			topo_data.data.combos.forEach((combo: any, i: any) => {
			graph.collapseCombo(combo['id']);
			})   
			init_collapse = true   
		}

		graph.updateCombos();
	}

}, {deep: true})

watch(() => props.time_interval, (newdata) => {
    switch (newdata) {
      case '关闭':
        clearInterval(timer)
        break;
      default:
        switch (newdata) {
          case '5s':
            time_interval_num = 5000
            break;
          case '10s':
            time_interval_num = 10000
            break;
          case '15s':
            time_interval_num = 15000
            break;
          case '1m':
            time_interval_num = 60000
            break;
          case '5m':
            time_interval_num = 300000
            break;
        }

        try {
          if (timer != 0) {
            clearInterval(timer)
          }
          timer = setInterval(timerhandler, time_interval_num)
          // console.log('timer: ', timer)
        } catch (error) {
          console.error(error)
        }
    }
})

</script>

<style scoped>

.container {
  width: 100%;
  height: 100%;
  background-color: white;
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
  background-color: #67e1e3ce;
}

</style>
