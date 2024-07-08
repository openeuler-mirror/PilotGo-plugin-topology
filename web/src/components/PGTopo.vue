<template>
  <div id="topo-container" class="container">
    <div class="topo_legend">
      <p><img width="20" src="../assets/icon/server_blue.png" alt=""><span>主机</span></p>
      <p><img width="20" src="../assets/icon/process.png" alt=""><span>进程</span></p>
      <p><img width="20" src="../assets/icon/net.png" alt=""><span>网络</span></p>
      <p><img width="20" src="../assets/icon/resource.png" alt=""><span>资源</span></p>
    </div>
  </div>
</template>

<script setup lang="ts">
import G6, { Graph } from '@antv/g6';
import type { ICombo, INode } from '@antv/g6';
import { ref, reactive, onMounted, watch, nextTick } from "vue";
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

let topo_container: HTMLElement;

let graph: Graph;
let topo_data = reactive<any>({});
let topo_type = ref('comb');

let combo_positions: Map<string, number[]>;
let combo_color: Map<string, string[]>;
let combo_collapse_status: Map<string, boolean>;
let zoom_before: number;
let node_position_process: Map<string, number[]>;
let node_position_host: Map<string, number[]>;

const topoW = ref(0);
const topoH = ref(0);

onMounted(() => {
  topo_container = document.getElementById("topo-container")!;
  topoW.value = topo_container.clientWidth;
  topoH.value = topo_container.clientHeight;
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

  // 初始化combo收缩/展开状态 combo_collapse_status
  if (!combo_collapse_status) {
    combo_collapse_status = new Map<string, boolean>();
    topo_data.combos.forEach((combo: any, _i: any) => {
      combo_collapse_status.set(combo.id as string, true);
    });
  }
  // 初始化combo颜色 combo_color
  if (!combo_color || combo_color.size < topo_data.combos.length) {
    combo_color = new Map<string, string[]>();
    topo_data.combos.forEach((combo: any, i: number) => {
      combo_color.set(combo.id, [colorSets[i].mainStroke as string, colorSets[i].mainFill as string]);
    });
  }

  topo_data.combos.forEach((combo: any, _i: number) => {
    const colors = combo_color.get(combo.id);
    if (colors) {
      combo.style = {
        stroke: colors[0],
        fill: colors[1],
        opacity: 0.8
      };
    }
    combo.collapsed = true; // 图固定关键参数，原因未知
  })
}

// 定义节点右键菜单
const menu = new G6.Menu({
  offsetX: 6,
  offsetY: 10,
  itemTypes: ['node'],
  shouldBegin(_e: any) {
    if (_e.item._cfg.model.Type === 'host') {
      return false;
    }
    return true;
  },
  getContent(_e) {
    const outDiv = document.createElement('div');
    outDiv.style.width = '80px';
    outDiv.innerHTML = `
    <span style="font-size:14px; cursor:pointer;">
      查看日志
    </span>
    `
    return outDiv;
  },
  // _target：界面元素，item：节点内容
  handleMenuClick(_target, item: any) {
    if (item._cfg) {
      let host_name = '' as any; let process_name = '' as any;
      let node_type = item._cfg.model.Type;
      if (node_type === 'process') {
        process_name = item._cfg.model.name;
        host_name = graph.getNeighbors(item._cfg.id!)[0]._cfg!.model!.metrics?.Hostname;
      } else if (node_type === 'host') {
        process_name = host_name = item._cfg.model!.name;
      }

      useTopoStore().node_click_info.node_id = item._cfg.id!;
      useTopoStore().node_click_info.host_name = host_name;
      useTopoStore().node_click_info.process_name = process_name;
    }
  },
});

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
      plugins: [menu],
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
  graph.on('node:click', (e) => {
    graph.getNodes().forEach((node) => {
      graph.clearItemStates(node);
    });
    const nodeItem = e.item;
    if (nodeItem) {
      graph.setItemState(nodeItem, 'click', true);
      // 抽屉组件展示的节点数据
      let selected_node = nodeItem._cfg;
      if (selected_node && topo_type.value === 'comb') {
        useTopoStore().nodeData = selected_node.model;
      } else if (selected_node && selected_node.model) {
        useTopoStore().nodeData = selected_node.model.node;
      }
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
    refreshDragedNodePosition(e);
  });

  graph.on('node:dragend', (e) => {
    node_position_process.set(e.item?.getModel().id as string, [e.item?.getModel().x as number, e.item?.getModel().y as number])
    if (e.item?.getModel().id === '3df830f0-cc71-48b4-8693-cd36098cc6d1_process_147139') {
      console.log("node drag end: ", node_position_process.get(e.item?.getModel().id as string));
    }
  });

  graph.on('combo:dragend', (e) => {
    combo_positions.set(e.item?.getModel().id as string, [e.item?.getModel().x as number, e.item?.getModel().y as number]);
  });

  // 保存combo collapse expand状态
  graph.on('combo:dblclick', (e) => {
    combo_collapse_status.set(e.item?.getID() as string, e.item?.getModel().collapsed as boolean);
  });

  // 解决拖动产生残影问题
  graph.get('canvas').set('localRefresh', false);

  graph.on('wheelzoom', (_e) => {
    zoom_before = graph.getZoom();
  });

  // 渲染前的操作
  graph.on('beforelayout', () => {

    // 重置combo坐标
    if (combo_positions) {
      graph.getCombos().forEach((combo: ICombo, _i: any) => {
        const position = combo_positions.get(combo.getModel().id as string)
        if (position) {
          combo.updatePosition({ x: position[0], y: position[1] });
        }
      });
    } else {
      combo_positions = new Map<string, number[]>();
    }

    // 重置process node坐标
    if (node_position_process) {
      graph.getNodes().forEach((node: any, _i: number) => {
        const process_position = node_position_process.get(node.id);
        if (node.getModel().Type === 'process' && process_position) {
          node.updatePosition({ x: process_position[0], y: process_position[1] })
        }
      });
    } else {
      // 初始化node_position_process
      node_position_process = new Map<string, number[]>();
    }

    if (node_position_host) {
      graph.getNodes().forEach((node: any, _i: number) => {
        const host_position = node_position_host.get(node.id);
        if (node.getModel().Type === 'host' && host_position) {
          node.updatePosition({ x: host_position[0], y: host_position[1] })
        }
      });
    } else {
      node_position_host = new Map<string, number[]>();
    }
  });

  // 渲染后的操作
  graph.on('afterlayout', () => {
    saveCombosPosition();

    // 恢复combo收缩/展开状态
    graph.getCombos().forEach((combo: ICombo, _i: any) => {
      if (combo_collapse_status.get(combo.getModel().id as string) === true) {
        graph.collapseCombo(combo.getModel().id as string);
      } else if (combo_collapse_status.get(combo.getModel().id as string) === false) {
        graph.expandCombo(combo.getModel().id as string);
      }
    });

    saveNodePosition();
  });

  graph.data(data);
  graph.render();


  while (!topo_container) {
    setTimeout(() => { }, 50);
  }

  if (typeof window !== 'undefined') {
    window.onresize = () => {
      graph.changeSize(topo_container.clientWidth, topo_container.clientHeight)
      graph.fitCenter()
    }
  }
}

function refreshDragedNodePosition(e: any) {
  const model = e.item.get('model');
  model.fx = e.x;
  model.fy = e.y;
}

// 保存当前combo坐标
function saveCombosPosition() {
  graph.getCombos().forEach((combo: ICombo, _i: number) => {
    const { x, y, id } = combo.getModel();
    if (x && y && id) {
      combo_positions.set(id, [x, y]);
    }
  });
  // console.log(combo_positions)
}

function saveNodePosition() {
  graph.getNodes().forEach((node: INode, _i: number) => {
    const { x, y, id } = node.getModel();
    if (x && y && id) {
      switch (node.getModel().Type) {
        case 'host':
          node_position_host.set(id, [x, y]);
          break;
        case 'process':
          node_position_process.set(id, [x, y]);
          break;
        default:
          break;
      }
    }

  });
  // console.log(node_position_host, node_position_process)
}

function changeZoom(zoom: number) {
  const point_position = graph.getPointByCanvas(graph.getWidth() / 2, graph.getHeight() / 2);
  graph.zoom(zoom, point_position, false);
  // graph.zoom(zoom);
}

watch(() => useTopoStore().topo_data, (new_topo_data) => {
  nextTick(() => {
    // 数据处理入口
    topo_type.value = 'comb';
    let topo_data_raw = JSON.parse(JSON.stringify(new_topo_data));
    if (topo_data_raw.tree) {
      topo_type.value = 'tree';
      initGraph(topo_data_raw.tree);
    } else if (topo_data_raw.nodes && !graph) {
      updateTopoData(topo_data_raw);
      initGraph(topo_data);
    } else if (topo_data_raw.nodes) {
      updateTopoData(topo_data_raw);
      graph.changeData(topo_data);
    }

  });

}, { immediate: true, deep: true })

// 设置graph_mode
watch(() => props.graph_mode, (newData) => {
  if (newData) {
    graph.setMode(newData);
  }
})

</script>

<style scoped lang="scss">
.topo_legend {
  position: absolute;
  top: 10;
  width: 300px;
  height: 60px;
  display: flex;
  align-items: center;

  p {
    width: 25%;
    display: flex;
    align-items: center;

    span {
      color: #666;
      padding-left: 2px;
    }
  }
}

.container {
  width: 100%;
  background-color: white;
}
</style>
