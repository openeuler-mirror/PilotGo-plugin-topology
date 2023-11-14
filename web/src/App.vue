<template>
  <div class="top_area_div">
    <div class="top_area_inner_div">
      <el-dropdown>
        <span class="dropdown">
          <el-icon><Menu /></el-icon>业务<el-icon class="el-icon--right"><arrow-down /></el-icon>
        </span>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item @click="switch_multi_topo">全局网络拓扑</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>

      <el-dropdown>
        <span class="dropdown">
          <el-icon><Monitor /></el-icon>机器<el-icon class="el-icon--right"><arrow-down /></el-icon>
        </span>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item @click="switch_single_topo(node)" v-for="node in node_list">{{ node.id }}</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>

      <el-dropdown>
        <span class="dropdown">
          <el-icon><Setting /></el-icon>设置<el-icon class="el-icon--right"><arrow-down /></el-icon>
        </span>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item>编辑业务</el-dropdown-item>
            <el-dropdown-item>1</el-dropdown-item>
            <el-dropdown-item>1</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>

      <div :style="{ display: 'flex', 'align-items': 'center' }">
        <span class="dropdown">
            <el-icon><Setting /></el-icon>时间<el-icon class="el-icon--right"></el-icon>
        </span>
        <el-date-picker v-model="dateRange" type="datetimerange" :shortcuts="pickerOptions" range-separator="to"
          start-placeholder="开始日期" end-placeholder="结束日期" @change="changeDate">
        </el-date-picker>
      </div>

      <el-dropdown>
        <span class="dropdown">
          <el-icon><Aim /></el-icon>时间间隔<el-icon class="el-icon--right"><arrow-down /></el-icon>
        </span>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item v-for="interval in interval_list">{{ interval }}</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
  </div>
  <RouterView />

</template>

<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router'
import { ref, reactive, onMounted } from "vue";
import { useRouter } from "vue-router";
import { topo } from '@/request/api';
import { useLayoutStore } from '@/stores/charts';
import { pickerOptions } from '@/utils/datePicker';

const node_list = reactive<any>([])
const time_list = reactive<any>(["1699595594", "1699595668", "1699595751"])
const interval_list = reactive<any>(["5s", "15s", "1m", "5m"])

let dateRange = ref([new Date() as any - 2 * 60 * 60 * 1000, new Date() as any - 0])
const startTime = ref(0);
const endTime = ref(0);
startTime.value = (new Date() as any) / 1000 - 60 * 60 * 2;
endTime.value = (new Date() as any) / 1000;

const layoutStore = useLayoutStore();
let layout = reactive(layoutStore.layout_option);

const router = useRouter()

onMounted(async () => {
  try {
    updateNodeList()
  } catch (error) {
    console.error(error)
  }

  router.push("/cluster")
})

async function updateNodeList() {
  //ttcode
  // const data = {
				// 	"code":  0,
				// 	"error": null,
				// 	"data": 
							// 		{"agentlist": {"070cb0b4-c415-4b6a-843b-efc51cff6b76": "10.44.55.66:9992"}}
        //   }
	
  const data = await topo.host_list()
  // console.log(data);
  for (let key in data.data.agentlist) {
    node_list.push({
      id: key,
    })
  };

}

function switch_multi_topo() {
  router.push("/cluster")
}

function switch_single_topo(node: any) {
  router.push({
    path: "/node",
    query: {
      uuid: node.id
    }
  })
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

</script>

<style scoped>
header {
  line-height: 1.5;
  max-height: 100vh;
  place-items: center;
  padding-right: calc(var(--section-gap) / 2);
}

.top_area_div {
  width: 100%;
  height: 5%;
  position: relative;
  display: flex;
  justify-content:center;
  background-color: #cfcaca;
}

.top_area_inner_div {
  display: flex;
  justify-content: space-between;

}

.top_area_inner_div > div {
  margin-right: 20px;
}

.top_area_inner_div > div:last-child {
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
</style>
