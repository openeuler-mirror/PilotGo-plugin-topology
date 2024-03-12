<template>

  <div class="topoContaint">
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
import { onMounted, ref, watchEffect } from 'vue';
import PGTopo from '@/components/PGTopo.vue';
// import topodata from '@/utils/test.json';
import nodeDetail from './nodeDetail.vue';
import { getCustomTopo, getTopoData, getUuidTopo } from "@/request/api";
import { useTopoStore } from '@/stores/topo';
import { useConfigStore } from '@/stores/config';
import router from '@/router';

const graphMode = ref('default');
const timeInterval = ref('10');
const showTopo = ref(false)

onMounted(() => {
  useTopoStore().topo_type = 'comb';
})

const goBack = () => {
  router.push('topoList');
}

watchEffect(async () => {
  let requst_type = useConfigStore().topo_request.type;
  let requst_id = useConfigStore().topo_request.id;
  let topoData = {};
  switch (requst_type) {
    case 'custom':
      await getCustomTopo({ id: requst_id as number }).then(res => {
        if (res.data.code === 200) {
          topoData = res.data.data;
        }
      })
      break;
    case 'single':
      useTopoStore().topo_type = 'tree';
      await getUuidTopo({ uuid: requst_id as string }).then(res => {
        if (res.data.code === 200) {
          topoData = res.data.data;
        }
      })
      break;

    default:
      await getTopoData().then(res => {
        if (res.data.code === 200) {
          topoData = res.data.data;
        }
      })
      break;
  }
  useTopoStore().topo_data = topoData;
  showTopo.value = true;
})

</script>

<style scoped lang="scss">
.topoContaint {
  width: 96%;
  height: 100%;

  margin: 0 auto;
}
</style>