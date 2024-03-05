import { ref, computed } from 'vue'
import { defineStore } from 'pinia'

export const useTopoStore = defineStore('topo', () => {
  const topo_type = ref('');
  const topo_data = ref({} as any);
  const nodeData = ref({} as any);
  return { topo_type, nodeData, topo_data }
})
