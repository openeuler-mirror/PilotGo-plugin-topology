/* 
 * Copyright (c) KylinSoft  Co., Ltd. 2024.All rights reserved.
 * PilotGo-plugin-topology licensed under the Mulan Permissive Software License, Version 2. 
 * See LICENSE file for more details.
 * Author: zhao_zhen_fang <zhaozhenfang@kylinos.cn>
 * Date: Tue Nov 14 16:56:53 2023 +0800
 */
import request from './request'
// 获取指标列表
export function getPromRules() {
  return request({
    url: '/plugin/prometheus/api/targets',
    method: 'get',
  })
}

// 获取prome某一时间点的数据
export function getPromeCurrent(data: object) {
  return request({
    url: '/plugin/prometheus/api/query',
    method: 'get',
    params: data
  })
}

// 获取prome某一时间段的数据
export function getPromeRange(data: object) {
  return request({
    url: '/plugin/prometheus/api/query_range',
    method: 'get',
    params: data
  })
}

// 获取监控主机ip
export function getMacIp() {
  return request({
    url: '/plugin_manage/info',
    method: 'get',
  })
}