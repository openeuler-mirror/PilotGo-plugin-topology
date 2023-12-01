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
        <el-dropdown :style="{ 'position': 'fixed', 'margin-top': '-12px', 'margin-left': '70px' }">
          <span class="dropdown">
            <el-icon class="el-icon--right" ><arrow-down /></el-icon>
          </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item @click="switchSingleTopo(node)" v-for="node in node_list">{{ node.id }}</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </el-tab-pane>
      

      <el-tab-pane label="设置" name="third">
        <el-dropdown :style="{ 'position': 'fixed', 'margin-top': '-12px', 'margin-left': '138px' }">
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

      <el-tab-pane :label="interactive_mode" name="fifth">
        <el-dropdown :style="{ 'position': 'fixed', 'margin-top': '-12px', 'margin-left': '330px' }">
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
        <el-dropdown :style="{ 'position': 'fixed', 'margin-top': '-12px', 'margin-left': '426px' }">
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
    <!-- <div>{{ '业务: ' + current_topo }}</div>
    <div>{{ '时间间隔: ' + time_interval }}</div> -->
    <div class="regional_table">
    <el-table :data="cart_table" style="width: 100%" :show-header="false" size="small" :row-style="{height: '0px'}">
      <el-table-column prop="name" label="name" width="70" />
      <el-table-column prop="content" label="content" width="100" />
    </el-table>
    </div>
  </el-card>
</div>

<RouterView :graph_mode="graph_mode" :time_interval="time_interval"/>

</template>

<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router'
import { ref, reactive, onMounted } from "vue";
import { useRouter } from "vue-router";
import { topo } from '@/request/api';
import { type TabsPaneContext } from 'element-plus'

const interactive_mode = ref('全局交互模式')
const appearance_mode = ref('亮色')
const graph_mode = ref('mixmode')
const activename = ref('first')
let agent_list = reactive<any>({})
const node_list = reactive<any>([])
const time_interval = ref('关闭')
const current_topo = ref('全局网络拓扑')
const interval_list = reactive(["关闭", "5s", "10s", "15s", "1m", "5m"])
const cart_table = reactive([
  {
    'name': '业务',
    'content': current_topo
  },
  {
    'name': '时间间隔',
    'content': time_interval
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
  const data = {
				"code":  0,
				"error": null,
				"data": 
							{"agentlist": {"070cb0b4-c415-4b6a-843b-efc51cff6b76": "10.44.55.66:9992"}}
        }
  // const data = await topo.host_list()
  
  agent_list = data.data.agentlist
  for (let key in agent_list) {
    node_list.push({
      id: key,
    })
  };

}

function switchMultiTopo() {
  current_topo.value = '全局网络拓扑'
  router.push("/cluster")
}

function switchSingleTopo(node: any) {
  current_topo.value = agent_list[node.id].split(':')[0]
  router.push({
    path: "/node",
    query: {
      uuid: node.id
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
  /* line-height: 1.5;
  max-height: 100vh;
  place-items: center;
  padding-right: calc(var(--section-gap) / 2); */

  width: 100%;
  /* height: 5%; */
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
  /* height: calc(100% - 123px); */
  height: 100%;
  /* padding: 50px; */
  color: #c71e48;
  /* font-size: large;
  font-weight: 600; */

}
.el-tabs__header {
  margin: 0px;
}

.el-tabs__item{
  font-size: 18px;
  /* margin-top: -10px; */
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
  width: 180px;
}

.box-card {
  position: absolute;
  height: 100%;
  width: 100%;
  background-color: rgba(207, 210, 21, 0.1);
  font-size: 14px;
}

.el-card /deep/ .el-card__body {
  padding: 5px !important;
}

.regional_table /deep/ .el-table, .regional_table /deep/ .el-table__expanded-cell {
  background-color: transparent;
  color: #606266;
}

.regional_table /deep/ .el-table th, .regional_table /deep/ .el-table tr, .regional_table /deep/ .el-table td {
  background-color: transparent !important;
  color: #606266;
}
</style>
