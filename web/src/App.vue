<template>
  <header>
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
        <el-icon><MagicStick /></el-icon>模式<el-icon class="el-icon--right"><arrow-down /></el-icon>
      </span>
      <template #dropdown>
        <el-dropdown-menu>
          <el-dropdown-item >亮色</el-dropdown-item>
          <el-dropdown-item >黑暗</el-dropdown-item>
        </el-dropdown-menu>
      </template>
    </el-dropdown>
</header>
  <RouterView/>

</template>

<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router'
import { ref, reactive, onMounted } from "vue";
import { useRouter } from "vue-router";
import { topo } from '@/request/api';

const node_list = reactive<any>([])
const interval_list = reactive<any>(["关闭", "5s", "10s", "15s", "1m", "5m"])

const startTime = ref(0);
const endTime = ref(0);
startTime.value = (new Date() as any) / 1000 - 60 * 60 * 2;
endTime.value = (new Date() as any) / 1000;

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

</script>

<style scoped>
header {
  /* line-height: 1.5;
  max-height: 100vh;
  place-items: center;
  padding-right: calc(var(--section-gap) / 2); */

  width: 100%;
  height: 5%;
  position: relative;
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

</style>
