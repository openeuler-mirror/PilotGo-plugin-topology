<template>
  <div class="topoContaint">
    <!-- 展示topo图 -->
    <PGTopo v-model:selectedNode="selectedNode" :topo_data="topo_data" :graph_mode="graphMode"
      :time_interval="timeInterval" />
    <!-- 嵌套抽屉组件展示数据 -->
    <nodeDetail :display_drawer="drawer.display" :node="selectedNode.model" />
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue';
import PGTopo from '@/components/PGTopo.vue';
// import topodata from '@/utils/test.json';
import nodeDetail from './nodeDetail.vue';
import { getCustomTopo, getTopoData, getUuidTopo } from "@/request/api";
import { useTopoStore } from '@/stores/topo';

const graphMode = ref('default');
const timeInterval = ref('10');
const topo_data = ref({});
let selectedNode = reactive({ model: {} });

const drawer = reactive({
  display: false,
  title: ''
})
onMounted(async () => {
  let requst_type = history.state.type;
  let topoData = {};
  switch (requst_type) {
    case 'custom':
      await getCustomTopo({ id: history.state.id }).then(res => {
        if (res.data.code === 200) {
          topoData = res.data.data;
        }
      })
      break;
    case 'single':
      await getUuidTopo({ uuid: history.state.id }).then(res => {
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
})

</script>

<style scoped lang="scss">
.topoContaint {
  width: 96%;
  height: 100%;

  margin: 0 auto;
}
</style>