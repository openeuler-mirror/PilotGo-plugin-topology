/* 
 * Copyright (c) KylinSoft  Co., Ltd. 2024.All rights reserved.
 * PilotGo-plugin-topology licensed under the Mulan Permissive Software License, Version 2. 
 * See LICENSE file for more details.
 * Author: zhaozhenfang <zhaozhenfang@kylinos.cn>
 * Date: Thu Mar 7 16:25:33 2024 +0800
 */
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
