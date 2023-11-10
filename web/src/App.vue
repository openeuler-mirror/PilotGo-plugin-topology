<template>
  <div class="top_area">
    <el-dropdown>
      <span class="el-dropdown-link">
        <el-icon><Menu /></el-icon>拓扑<el-icon class="el-icon--right"><arrow-down /></el-icon>
      </span>
      <template #dropdown>
        <el-dropdown-menu>
          <el-dropdown-item @click="switch_multi_topo">业务集群网络拓扑</el-dropdown-item>
          <el-dropdown-item @click="switch_single_topo">单机拓扑</el-dropdown-item>
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
import { More, Menu } from '@element-plus/icons-vue';

const router = useRouter()

function switch_multi_topo() {
  router.push("/cluster")
}

function switch_single_topo() {
  router.push("/node")
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

  .el-dropdown-link {
    font-size: medium;
    cursor: pointer;
    color:rgb(79, 104, 104);
    display: flex;

    align-items: center;
    margin-left: 5px;
    margin-top: 10px;
  }
}
</style>
