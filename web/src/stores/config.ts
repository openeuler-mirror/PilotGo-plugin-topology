import { ref, computed, reactive } from 'vue'
import { defineStore } from 'pinia'
interface TopoRequest {
    type: string;
    id: string | number;
  }
export const useConfigStore = defineStore('config', {
  state: () => ({
    topo_config: {} as any,
    topo_request:{
      type: '',
      id:''
    } as TopoRequest
  }),
  persist:true
})
