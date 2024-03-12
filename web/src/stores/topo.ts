import { ref } from 'vue'
import { defineStore } from 'pinia'

export const useTopoStore = defineStore('topo', () => {

  const topo_data = ref({} as any);
  const nodeData = ref({} as any);
  return {nodeData, topo_data }
})
