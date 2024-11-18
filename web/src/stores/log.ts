import { ref } from 'vue'
import { defineStore } from 'pinia'
interface LogSearchList {
    ip:string,
    timeRange:[Date,Date];
    level:string;
    service:{label:string,value:string};
    realTime:boolean;
}
export const useLogStore = defineStore('log', () => {
  const search_list = ref([] as LogSearchList[]);
  const ws_isOpen = ref(false);
  const clientId = ref(parseInt(Math.random() * 100000+'')); // websocket标识id，初始化为随机数
  const updateLogList = (param:any) => {
    let ip_index = search_list.value.findIndex(item => item.ip === param.ip);
    ip_index !== -1 ? search_list.value[ip_index] = param : search_list.value.push(param);
  }

  const $reset = () => {
    search_list.value = [];
    ws_isOpen.value = false;
  }
  return {clientId,ws_isOpen,search_list,updateLogList,$reset}
})