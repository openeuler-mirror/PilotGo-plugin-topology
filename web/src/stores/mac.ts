/* 
 * Copyright (c) KylinSoft  Co., Ltd. 2024.All rights reserved.
 * PilotGo-plugin-topology licensed under the Mulan Permissive Software License, Version 2. 
 * See LICENSE file for more details.
 * Author: Wangjunqi123 <wangjunqi@kylinos.cn>
 * Date: Tue Nov 14 16:41:40 2023 +0800
 */
import { defineStore } from 'pinia';

export const useMacStore = defineStore('mac', {
  state: () => {
    return {
      macIp: '',
    };
  },
  getters: {
    newIp(state) {
      return state.macIp.length > 0 ? state.macIp.split(':')[0] : '';
    },
  },
  actions: {
    setMacIp(ip: string) {
      this.macIp = ip;
      this.newIp;
    }
  }
});
