/* 
 * Copyright (c) KylinSoft  Co., Ltd. 2024.All rights reserved.
 * PilotGo-plugin-topology licensed under the Mulan Permissive Software License, Version 2. 
 * See LICENSE file for more details.
 * Author: zhao_zhen_fang <zhaozhenfang@kylinos.cn>
 * Date: Mon Oct 9 11:19:00 2023 +0800
 */
import axios from 'axios';
import { type AxiosInstance } from 'axios';
import router from '@/router';

// 1.创建axios实例
const request: AxiosInstance = axios.create({
  baseURL: '',
  timeout: 15000,
  headers: {
    // 设置后端需要的传参类型
    'Content-Type': 'application/json',
    token: '',
    'X-Requested-With': 'XMLHttpRequest',
  },
});

// 2.1添加请求拦截器
request.interceptors.request.use(
  (config) => {
    return config;
  },
  (error) => {
    return Promise.reject(error);
  },
);

// 2.2添加响应拦截器
request.interceptors.response.use(
  (response: any) => {
    if (response.data && response.data.code == '401') {
      router.push('/');
    } else {
      return response;
    }
  },
  (error) => {
    if (error.response) {
      switch (error.response.status) {
        case 401:
          router.push('/');
      }
      return Promise.reject(error.response.data);
    }
  },
);

export default request;
