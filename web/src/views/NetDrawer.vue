<template>
    <!-- 外层抽屉组件 -->
    <el-drawer class="drawer" v-model="net_drawer" :with-header="false" direction="rtl" size="480px" :before-close="handleClose" :modal="true" :append-to-body="false">
      <el-scrollbar class="drawer_head_div">
        <div v-for="(tag, i) in tags" :style="{ 'display': 'flex', 'margin-bottom': '5px' }">
          <span class="tag" :style="{ 'background-color': tags_color[i % tags_color.length] }">{{ tag }}</span>
        </div>
      </el-scrollbar>
  
      <div class="drawer_body_div">
        <grid-layout :col-num="3" :is-draggable="grid.draggable" :is-resizable="grid.resizable" :layout.sync="layout"
        :row-height="100" :use-css-transforms="true" :vertical-compact="true" :responsive="true">
          <template v-for="(item, i) in layout">
            <grid-item :key="i" :h="item.h" :i="item.i" :static="item.static" :w="item.w" :x="item.x" :y="item.y"
              :min-w="2" :min-h="2" @resize="SizeAutoChange(item.i, item.query.isChart)" @resized="SizeAutoChange"
              drag-allow-from=".drag" drag-ignore-from=".noDrag" v-if="item.display">
              <div class="drag">
                <span class="drag-title">{{ item.title }}</span>
              </div>
              <div class="noDrag">
                <MyEcharts :query="item.query" :startTime="startTime"
                  :endTime="endTime" :style="{ 'width': '100%', 'height': '100%'}">
                </MyEcharts>
              </div>
            </grid-item>
          </template>
        </grid-layout>
      </div>
  
      <!-- 嵌套抽屉组件 -->
      <div class="nested_metric_drawer_div">
        <el-drawer v-model="nested_metric_drawer" :with-header="false" :append-to-body="true" size="350px">
          <el-table :data="table_data" stripe style="width: 100%">
            <el-table-column prop="name" label="属性" />
            <el-table-column prop="value" label="值" />
          </el-table>
        </el-drawer>
      </div>
      <div class="nested_selectchart_drawer_div">
        <el-drawer v-model="nested_selectchart_drawer" :with-header="false" :append-to-body="true" size="190px">
          <el-checkbox v-for="item in layout" v-model="item.display" :label="item.title" size="large"/>
        </el-drawer>
      </div>
  
      <div class="drawer_top_div" :style="{ 'margin-top': '10px' }">
        <!-- 时间范围选择 -->
        <el-date-picker  v-model="dateRange" type="datetimerange" :shortcuts="pickerOptions" range-separator="至"
          start-placeholder="开始日期" end-placeholder="结束日期" @change="changeDate" size="small"
          :style="{ 'width': '285px', 'height': '30px', 'margin-right': '8px', 'border-radius': '10px', 'border': '1px groove rgb(223, 210, 210)' }">
        </el-date-picker>
        <!-- 按钮组 -->
        <el-button-group :style="{ 'margin-right': '22px' }">
          <!-- 指标数据 -->
          <el-button class="drawer_button" @click="nested_metric_drawer = true" :icon="More" size="default" circle="false" />
          <!-- 选择要显示的图表 -->
          <el-button class="drawer_button" @click="nested_selectchart_drawer = true" :icon="Platform" size="default" circle="false" />
          <!-- 加载本地的图表配置文件 -->
          <el-button class="drawer_button" @click="config_drawer_inner = true" :icon="Files" size="default" circle="false" />
          <!-- 保存图标显示配置 -->
          <el-button class="drawer_button" @click="config_drawer_inner = true" :icon="Collection" size="default" circle="false" />
        </el-button-group>
      </div>
      
    </el-drawer>
  </template>
  
  <script setup lang="ts">
  import { ref, reactive, onMounted, watch } from "vue";
  import { More, Platform, Files, Collection } from '@element-plus/icons-vue';
  import { useLayoutStore } from '@/stores/charts';
  import MyEcharts from '@/views/MyEcharts.vue';
  import { pickerOptions } from '@/utils/datePicker';
  import { useMacStore } from '@/stores/mac';
  
  let tags = reactive<any>([])
  let table_data = reactive<any>([])
  let node = reactive<any>({})
  const net_drawer = ref(false)
  const props = defineProps({
      net_drawer: {
          type: Boolean,
          default: false,
          required: true,
      },
      node: {
          type: Object,
          default: {},
          requried: true,
      },
  })
  
  const emit = defineEmits(['update-statu'])
  
  let nested_metric_drawer = ref(false)
  let nested_selectchart_drawer = ref(false)
  let config_drawer_inner = ref(false)
  const chart = ref([] as any);
  
  let dateRange = ref([new Date() as any - 2 * 60 * 60 * 1000, new Date() as any - 0])
  const startTime = ref(0);
  const endTime = ref(0);
  startTime.value = (new Date() as any) / 1000 - 60 * 60 * 2;
  endTime.value = (new Date() as any) / 1000;
  
  const layoutStore = useLayoutStore();
  let layout = reactive(layoutStore.net_layout_option);
  
  const grid = reactive({
    draggable: true,
    resizable: true,
    responsive: true,
  });
  
  const tags_color: string[] = [
    'rgb(86, 148, 128)',
    'rgb(218, 113, 148)',
    'rgb(255, 196, 84)',
    'rgb(76, 142, 218)',
    'rgb(236, 181, 201)',
    'rgb(141, 204, 147)',
    'rgb(217, 200, 174)',
    'rgb(241, 102, 103)'
  ];
  
  onMounted(() => {
      net_drawer.value = props.net_drawer;
      node = props.node
      
  })
  
  watch(() => props.net_drawer, (new_data) => {
      net_drawer.value = new_data as any;
      node = props.node
  
      if (net_drawer.value) { 
          table_data = [];
          let metrics = node.model.metrics;
          for (let key in metrics) {
              table_data.push({
              name: key,
              value: metrics[key],
              })
          };
  
          tags = [];
          for (let i in node.model.tags) {
              tags.push(node.model.tags[i])
          };
      }
  
  }, {
    deep: true
  })
  
  function handleClose() {
      // net_drawer.value = false
      emit('update-statu')
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
  
  // echarts大小随grid改变
  const SizeAutoChange = (i: string, isChart?: boolean) => {
    if (isChart) {
      chart.value[i].resize();
    }
  }
  </script>
  
  <style scoped>
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
  </style>