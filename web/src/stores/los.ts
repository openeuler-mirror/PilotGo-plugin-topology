import { ref } from 'vue'
import { defineStore } from 'pinia'
interface LogSearchObj {
    timeRange:[Date,Date];
    level:string;
    service:{label:string,value:string;}
}
interface LogSearchList {
    ip:string;
    search_obj:LogSearchObj;
}
export const useLogStore = defineStore('log', () => {
  const search_list = ref([] as LogSearchList[]);

  const updateLogList = (param:any) => {
    let ip_index = search_list.value.findIndex(item => item.ip === param.ip);
    ip_index !== -1 ? search_list.value[ip_index] = param : search_list.value.push(param);
  }

  const $reset = () => {
    search_list.value = [];
  }
  return {search_list,updateLogList,$reset}
})