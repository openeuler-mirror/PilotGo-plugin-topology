/* 
 * Copyright (c) KylinSoft  Co., Ltd. 2024.All rights reserved.
 * PilotGo-plugin-topology licensed under the Mulan Permissive Software License, Version 2. 
 * See LICENSE file for more details.
 * Author: zhaozhenfang <zhaozhenfang@kylinos.cn>
 * Date: Thu Feb 29 09:56:15 2024 +0800
 */
import G6 from '@antv/g6';
import machine_logo from "@/assets/icon/machine.png";

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

// 设置颜色
export const colorSets = G6.Util.getColorSetsBySubjectColors(
  subjectColors,
  backColor,
  theme,
  disableColor,
);

// 画布初始化配置
export const graphInitOptions = {
  groupByTypes: false,
  fitView: true, // 是否自适应画布
  fitViewPadding: 200, // 画布周围的留白px
  animate: true, // 是否开启动画效果
  zoom: 1,
  minZoom: 0.00000001, // 
  defaultNode: {
    size: 20,
    labelCfg: {
      position: "bottom",
      offset: 2,
    },
  },
  defaultCombo: {
    animate: false,
    // fixSize: 400,
    fixCollapseSize: 50,
    collapsed: true,
    // padding: 30,
    collapsedSubstituteIcon: {
      show: true,
      img: machine_logo,
      width: 50,
      height: 50
    },
    labelCfg: {
      position: 'top',
    },
  },
  // 边状态样式,暂定,缺tooltip
  edgeStateStyles: {
    // click: {
    //   stroke: '#0282FF',
    //   shadowBlur: 0,
    //   'text-shape': {
    //     fill: "#0282FF",
    //     fontWeight: 600,
    //   }
    // }
  },
  // 节点状态样式
  nodeStateStyles: {
    // 选中后样式
    // click: {
    //   fill: '#0282FF', // 填充色
    //   stroke: '#0282FF', // 节点描边颜色
    //   lineWidth: 1, // 描边宽度
    //   shadowColor: 'rgba(0,102,210,0.5)',
    //   'text-shape': {
    //     fill: "#0282FF"
    //   }
    // },
    // 悬浮后样式
    active: {
      fill: '#CDEEFF', // 填充色
      stroke: '#2EA1FF', // 节点描边颜色
      lineWidth: 1, // 描边宽度
      shadowColor: 'rgba(78,89,105,0.3)',
      'text-shape': {
        fill: "#0282FF",
        fontWeight: 500,
      }
    }
  },
  layout: {
    type: 'comboCombined',
    outerLayout: new G6.Layout['forceAtlas2']({
      gravity: 1,
      factor: 2,
      preventOverlap: true,
      linkDistance: (_edge: any, source: any, target: any) => {
        const nodeSize = ((source.size?.[0] || 30) + (target.size?.[0] || 30)) / 2;
        return Math.min(nodeSize * 1.5, 700);
      }
    }),
    innerLayout: new G6.Layout['concentric']({
      sweep: 6.28,
      preventOverlap: true,     
      nodeSize: 30,                    
      equidistant: false,      
      startAngle: 0,           
      clockwise: false,        
      maxLevelDiff: '0.1',        
      sortBy: 'layoutattr',   
      workerEnabled: true,    
    }), 
  },
  modes: {
    default: [
      'drag-canvas', 
      'zoom-canvas', 
      'drag-combo', 
      {
        type: 'collapse-expand-combo',
        trigger: 'dblclick',
        relayout: false, // 收缩展开后，不重新布局
      },
      'drag-node']
  },
}

export const graphTreeInitOptions = {
  modes: {
    default: ['drag-canvas', 'zoom-canvas', "click-select", "drag-node",
      {
        type: 'collapse-expand',
        onChange: function onChange(item: any, collapsed: any) {
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
}