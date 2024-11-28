/* 
 * Copyright (c) KylinSoft  Co., Ltd. 2024.All rights reserved.
 * PilotGo-plugin-topology licensed under the Mulan Permissive Software License, Version 2. 
 * See LICENSE file for more details.
 * Author: Wangjunqi123 <wangjunqi@kylinos.cn>
 * Date: Tue Nov 14 17:46:29 2023 +0800
 */
export let pickerOptions = [
  {
    text: '前1分钟',
    value: () => {
      const end = new Date()
      const start = new Date()
      start.setTime(start.getTime() - 60 * 1000)
      return [start, end]
    },
  },
  {
    text: '前30分钟',
    value: () => {
      const end = new Date()
      const start = new Date()
      start.setTime(start.getTime() - 60 * 30 * 1000)
      return [start, end]
    },
  },
  {
    text: '前1小时',
    value: () => {
      const end = new Date()
      const start = new Date()
      start.setTime(start.getTime() - 3600 * 1000)
      return [start, end]
    },
  },
  {
    text: '前1天',
    value: () => {
      const end = new Date()
      const start = new Date()
      start.setTime(start.getTime() - 3600 * 1000 * 24)
      return [start, end]
    },
  },
]