<template>
  <header>
    <el-tabs v-model="activename" @tab-click="handleClick" class="tabs">
      <el-tab-pane  label="业务" name="first" >
        <el-dropdown :style="{ 'position': 'fixed', 'margin-top': '-12px', 'margin-left': '1px' }">
          <span class="dropdown">
            <el-icon class="el-icon--right"><arrow-down /></el-icon>
          </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item @click="switchMultiTopo">全局网络拓扑</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </el-tab-pane>

      <el-tab-pane label="机器" name="second">
        <el-dropdown :style="{ 'position': 'fixed', 'margin-top': '-12px', 'margin-left': '69px' }">
          <span class="dropdown">
            <el-icon class="el-icon--right" ><arrow-down /></el-icon>
          </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item @click="switchSingleTopo(node)" v-for="node in node_list">{{ node }}</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </el-tab-pane>
      

      <el-tab-pane label="设置" name="third">
        <el-dropdown :style="{ 'position': 'fixed', 'margin-top': '-12px', 'margin-left': '137px' }">
          <span class="dropdown">
            <el-icon class="el-icon--right"><arrow-down /></el-icon>
          </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item>编辑业务</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </el-tab-pane>

      <el-tab-pane label="时间间隔" name="fourth">
        <el-dropdown :style="{ 'position': 'fixed', 'margin-top': '-12px', 'margin-left': '220px' }">
          <span class="dropdown">
            <el-icon class="el-icon--right"><arrow-down /></el-icon>
          </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item @click="changeTimeInterval(interval)" v-for="interval in interval_list">{{ interval }}</el-dropdown-item> 
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </el-tab-pane>

      <el-tab-pane label="交互模式" name="fifth">
        <el-dropdown :style="{ 'position': 'fixed', 'margin-top': '-12px', 'margin-left': '315px' }">
          <span class="dropdown">
            <el-icon class="el-icon--right"><arrow-down /></el-icon>
          </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item @click="changeInteractiveMode('default')">全局交互模式</el-dropdown-item>
              <el-dropdown-item @click="changeInteractiveMode('localmode')">单机交互模式</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </el-tab-pane>

      <el-tab-pane :label="appearance_mode" name="sixth">
        <el-dropdown :style="{ 'position': 'fixed', 'margin-top': '-12px', 'margin-left': '397px' }">
          <span class="dropdown">
            <el-icon class="el-icon--right"><arrow-down /></el-icon>
          </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item @click="changeAppearenceMode('亮色')">亮色外观模式</el-dropdown-item>
              <el-dropdown-item @click="changeAppearenceMode('暗色')">暗色外观模式</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </el-tab-pane>

  </el-tabs>
</header>

<div class="box_card_div">
  <el-card class="box-card" shadow="never">
    <div class="regional_table">
      <el-descriptions class="description_table" :column="1" size="default" border>
        <el-descriptions-item label="业务" label-class-name="description_table_label">{{ current_topo }}</el-descriptions-item>
        <el-descriptions-item label="时间间隔">{{ time_interval }}</el-descriptions-item>
        <el-descriptions-item label="交互模式">{{ interactive_mode }}</el-descriptions-item>
        <el-descriptions-item label="时间">{{ topo_time }}</el-descriptions-item>
      </el-descriptions>
    </div>
  </el-card>
</div>

<RouterView :graph_mode="graph_mode" :time_interval="time_interval" @update-topo-time="updateTopoTimeHandle"/>

</template>

<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router'
import { ref, reactive, onMounted, watch } from "vue";
import { useRouter } from "vue-router";
import { topo } from '@/request/api';
import { type TabsPaneContext } from 'element-plus'

const interactive_mode = ref('全局交互模式')
const appearance_mode = ref('亮色')
const graph_mode = ref('mixmode')
const current_topo = ref('全局网络拓扑')
const topo_time = ref('')
const activename = ref('first')

let agent_list_data: any
let agent_list = reactive<any>({})
let node_list = ref<string[]>([])

const time_interval = ref('关闭')
let time_interval_num: number = 0
let timer: number = 0
const interval_list = reactive(["关闭", "5s", "10s", "15s", "1m", "5m"])

const cart_table = reactive([
  {
    'name': '业务',
    'content': current_topo
  },
  {
    'name': '时间间隔',
    'content': time_interval
  },
  {
    'name': '交互模式',
    'content': interactive_mode
  },
  {
    'name': '时间',
    'content': topo_time
  }
])

const startTime = ref(0);
const endTime = ref(0);
startTime.value = (new Date() as any) / 1000 - 60 * 60 * 2;
endTime.value = (new Date() as any) / 1000;

const router = useRouter()

onMounted(async () => {
  try {
    updateHostList()
  } catch (error) {
    console.error(error)
  }

  router.push("/cluster")
})


async function updateHostList() {
  //ttcode
  // const agent_list_data = {
	// 			"code":  0,
	// 			"error": null,
	// 			"data": 
	// 						{"agentlist": {"070cb0b4-c415-4b6a-843b-efc51cff6b76": "10.44.55.66:9992"}}
  //       }
  agent_list_data = await topo.host_list()
  
  let temp_node_list: string[] = []

  agent_list = agent_list_data.data.agentlist
  for (let key in agent_list) {
    temp_node_list.push(key)
  };

  node_list.value = temp_node_list
}

function updateTopoTimeHandle(topotime: string) {
  topo_time.value = topotime
}

watch(() => time_interval.value, (newdata) => {
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
        timer = setInterval(updateHostList, time_interval_num)
        // console.log('timer: ', timer)
      } catch (error) {
        console.error(error)
      }
  }
})

function switchMultiTopo() {
  current_topo.value = '全局网络拓扑'
  router.push("/cluster")
}

function switchSingleTopo(node: string) {
  current_topo.value = agent_list[node].split(':')[0]
  router.push({
    path: "/node",
    query: {
      uuid: node
    }
  })
}

const handleClick = (tab: TabsPaneContext, event: Event) => {
  console.log(tab, event)
}

function changeInteractiveMode(mode: string) {
  graph_mode.value = mode

  switch (mode) {
    case 'default':
      interactive_mode.value = '全局交互模式'
      break;
    case 'localmode':
      interactive_mode.value = '单机交互模式'
      break;
  }
}

function changeAppearenceMode(mode: string) {
  appearance_mode.value = mode
}

function changeTimeInterval(interval: string) {
  time_interval.value = interval
}

</script>

<style scoped>
header {
  width: 100%;
  position: fixed;
  height: 38px;
  display: flex;
  justify-content: center;
  background-color: #cfcaca;
  
}

header > div {
  margin-right: 20px;
}

header > div:last-child {
  margin-right: 0;
}

.dropdown {
  font-size: medium;
  cursor: pointer;
  color:rgb(79, 104, 104);
  display:flex;
  outline-style: none;

  align-items: center;
  }

.tabs {
  position: relative;
  height: 100%;
  color: #c71e48;
}

.el-tabs__header {
  margin: 0px;
}

.el-tabs__item{
  font-size: 18px;
  color: rgb(96, 92, 92);

}

.el-tabs__item.is-active {
  color: rgb(6, 150, 176);
}

.box_card_div {
  position: fixed;
  top: 0;
  left: 0;
  margin-top: 40px;
  height: 150px;
  width: 225px;
}

.box-card {
  position: absolute;
  height: 100%;
  width: 100%;
  background-color: rgba(207, 210, 21, 0.1);
  font-size: 14px;
}

.el-card /deep/ .el-card__body {
  padding-top: 10px !important;
  padding-left: 10px !important;
  padding-right: 10px !important;
}

.regional_table /deep/ .el-descriptions {
  opacity: 0.8;
}

.regional_table /deep/ .el-descriptions__body .el-descriptions__table.is-bordered .el-descriptions__cell {
  background-color: rgba(207, 210, 21, 0.1);
  border: rgba(15, 5, 5, 0);
  padding-top: 2px;
  padding-left: 0px;
  padding-right: 5px;
  padding-bottom: 0px;
}

.regional_table /deep/ .el-descriptions, .regional_table /deep/ .el-descriptions__expanded-cell {
  color: #606266;
}

.regional_table /deep/ .el-descriptions th, .regional_table /deep/ .el-descriptions tr, .regional_table /deep/ .el-descriptions td {
  /* background-color: rgba(207, 210, 21, 0.1); */
  /* background-color: transparent !important; */
  color: #606266;
}                                         


</style>
