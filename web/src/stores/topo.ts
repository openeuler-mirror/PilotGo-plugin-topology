import { ref } from 'vue'
import { defineStore } from 'pinia'
interface nodeItem {
  host_name: string;
  node_id: string;
  process_name: string;
}
export const useTopoStore = defineStore('topo', () => {

  const topo_data = ref({} as any);
  const node_log_id = ref('');
  const node_click_info = ref({} as nodeItem) 
  const nodeData = ref({} as any);
  const edgeData = ref({} as any);

  const $reset = () => {
    node_click_info.value.host_name = '';
    node_click_info.value.node_id = '';
    node_click_info.value.process_name = '';
  }
  return {nodeData, node_log_id,node_click_info,topo_data, edgeData,$reset}
})
