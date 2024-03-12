import { ref, computed, reactive } from 'vue'
import { defineStore } from 'pinia'

export const useConfigStore = defineStore('config', () => {
  const topo_config = ref({} as any);
  interface TopoRequest {
    type: string;
    id: string | number;
  }
  const topo_request = reactive<TopoRequest>({
    type: '',
    id:''
  })
  return { topo_config,topo_request}
})
