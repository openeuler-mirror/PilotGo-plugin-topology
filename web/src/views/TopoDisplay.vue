<template>

  <div class="topoContaint" v-loading="loading">
    <el-page-header @back="goBack">
      <template #content>
        <span style="font-size: 14px;"> 拓扑图 </span>
      </template>
    </el-page-header>
    <!-- 展示topo图 -->
    <PGTopo style="width: 100%;height: 100%;" v-if="showTopo" :graph_mode="graphMode" :time_interval="timeInterval" />
    <!-- 嵌套抽屉组件展示数据 -->
    <nodeDetail />
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref, watch, watchEffect } from 'vue';
import PGTopo from '@/components/PGTopo.vue';
// import topodata from '@/utils/test.json';
import nodeDetail from './nodeDetail.vue';
import { getCustomTopo, getTopoData, getUuidTopo } from "@/request/api";
import { useTopoStore } from '@/stores/topo';
import { useConfigStore } from '@/stores/config';
import router from '@/router';

const graphMode = ref('default');
const timeInterval = ref('10');
const showTopo = ref(false);
const loading = ref(false);
onMounted(() => {
  loading.value = true;
})

const goBack = () => {
  router.push('topoList');
}

watchEffect(() => {
  let requst_type = useConfigStore().topo_request.type;
  let requst_id = useConfigStore().topo_request.id;
  let topoData = {};
  switch (requst_type) {
    case 'custom':
      getCustomTopo({ id: requst_id as number }).then(res => {
        if (res.data.code === 200) {
          topoData = res.data.data;
          useTopoStore().topo_data = topoData;
          showTopo.value = true;
          loading.value = false;
        }
      })
      break;
    case 'single':
      getUuidTopo({ uuid: requst_id as string }).then(res => {
        if (res.data.code === 200) {
          topoData = res.data.data;
          loading.value = false;
          showTopo.value = true;
          setTimeout(() => {
            useTopoStore().topo_data = topoData;
          }, 200)
        }
      })
      break;

    default:
      getTopoData().then(res => {
        if (res.data.code === 200) {
          topoData = res.data.data;
          useTopoStore().topo_data = topoData;
          showTopo.value = true;
          loading.value = false;
        }
      })
      break;
  }
})


</script>

<style scoped lang="scss">
.topoContaint {
  width: 96%;
  height: 100%;

  margin: 0 auto;
}
</style>