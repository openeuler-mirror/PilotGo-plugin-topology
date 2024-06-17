import { ref } from 'vue'
import { defineStore } from 'pinia'

export const useTopoStore = defineStore('topo', () => {

  const topo_data = ref({} as any);
  const node_log_id = ref('');
  const nodeData = ref({} as any);
  const edgeData = ref({} as any);
  return {nodeData, node_log_id,topo_data, edgeData}
})
