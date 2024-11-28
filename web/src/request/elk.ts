/* 
 * Copyright (c) KylinSoft  Co., Ltd. 2024.All rights reserved.
 * PilotGo-plugin-topology licensed under the Mulan Permissive Software License, Version 2. 
 * See LICENSE file for more details.
 * Author: zhaozhenfang <zhaozhenfang@kylinos.cn>
 * Date: Wed Jun 12 17:21:06 2024 +0800
 */
import request from './request';
// 请求elk主机日志信息
export function getELKLogData(data:any) {
  return request({
    url: '/plugin/elk/api/log_clusterhost_timeaxis_data',
    method: 'post',
    data
  })
}

// 请求elk进程日志信息
export function getELKProcessLogData(data:any) {
  return request({
    url: '/plugin/elk/api/log_hostprocess_timeaxis_data',
    method: 'post',
    data
  })
}

// 请求elk进程单个日志文件流
export function getELKProcessLogStream(data:any) {
  return request({
    url: '/plugin/elk/api/log_stream_data',
    method: 'post',
    data
  })
}