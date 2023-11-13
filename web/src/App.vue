<template>
  <div class="top_area">
    <el-dropdown>
      <span class="dropdown">
        <el-icon><Menu /></el-icon>拓扑<el-icon class="el-icon--right"><arrow-down /></el-icon>
      </span>
      <template #dropdown>
        <el-dropdown-menu>
          <el-dropdown-item @click="switch_multi_topo">业务集群网络拓扑</el-dropdown-item>
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
        <el-icon><Clock /></el-icon>时间<el-icon class="el-icon--right"><arrow-down /></el-icon>
      </span>
      <template #dropdown>
        <el-dropdown-menu>
          <el-dropdown-item v-for="time in time_list">{{ time }}</el-dropdown-item>
        </el-dropdown-menu>
      </template>
    </el-dropdown>

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

    <el-dropdown>
      <span class="dropdown">
        <el-icon><Setting /></el-icon>设置<el-icon class="el-icon--right"><arrow-down /></el-icon>
      </span>
      <template #dropdown>
        <el-dropdown-menu>
          <el-dropdown-item>1</el-dropdown-item>
          <el-dropdown-item>1</el-dropdown-item>
          <el-dropdown-item>1</el-dropdown-item>
        </el-dropdown-menu>
      </template>
    </el-dropdown>
  </div>
  <RouterView />

</template>

<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router'
import { ref, reactive, onMounted } from "vue";
import { useRouter } from "vue-router";
import { topo } from '@/request/api';

const node_list = reactive<any>([])
const time_list = reactive<any>(["1699595594", "1699595668", "1699595751"])
const interval_list = reactive<any>(["5s", "15s", "1m", "5m"])

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
  // router.push("/node")
  router.push({
    path: "/node",
    query: {
      uuid: node.id
    }
  })
}

</script>

<style scoped>
header {
  line-height: 1.5;
  max-height: 100vh;
  place-items: center;
  padding-right: calc(var(--section-gap) / 2);
}

.top_area {
  width: 100%;
  height: 5%;
  position: relative;
  background-color: #cfcaca;
}

.dropdown {
    font-size: small;
    cursor: pointer;
    color:rgb(79, 104, 104);
    display: flex;
    outline-style: none;

    align-items: center;
    margin-left: 20px;
    margin-top: 10px;
  }
</style>
