import { ref, computed, reactive } from 'vue'
import { defineStore } from 'pinia'
interface TopoRequest {
  type: string;
  batch_id: number;
    id: string | number;
  }
export const useConfigStore = defineStore('config', {
  state: () => ({
    topo_config: {} as any,
    topo_request:{
      type: '',
      batch_id: 0,
      id:''
    } as TopoRequest
  }),
  persist:true
})
