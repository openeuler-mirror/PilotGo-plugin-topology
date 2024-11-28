/* 
 * Copyright (c) KylinSoft  Co., Ltd. 2024.All rights reserved.
 * PilotGo-plugin-topology licensed under the Mulan Permissive Software License, Version 2. 
 * See LICENSE file for more details.
 * Author: Wangjunqi123 <wangjunqi@kylinos.cn>
 * Date: Mon Oct 9 11:19:00 2023 +0800
 */
import { useLogStore } from '@/stores/log';
import request from './request';
let baseURL = '/plugin/topology';
// 请求批次信息
export function getBatchList() {
  return request({
    url: baseURL+'/api/batch_list',
    method: 'get',
  })
}

// 请求批次详情
export function getBatchDetail(data: { batchId: number }) {
  return request({
    url: baseURL+'/api/batch_uuid',
    method: 'get',
    params: data
  })
}

// 请求拓扑配置列表
export function getConfList(data: {}) {
  return request({
    url: baseURL+'/api/custom_topo_list',
    method: 'get',
    params: data
  })
}

// 新增拓扑配置列表
export function addConfList(data: {}) {
  return request({
    url: baseURL+'/api/create_custom_topo',
    method: 'POST',
    data
  })
}

// 更新拓扑配置列表
export function updateConfList(data: {}) {
  return request({
    url: baseURL+'/api/update_custom_topo',
    method: 'put',
    data
  })
}


// 删除拓扑配置
export function delConfig(data: { ids: number[] }) {
  return request({
    url: baseURL+'/api/delete_custom_topo',
    method: 'delete',
    data: data
  })
}

// 请求某一个拓扑图数据
export function getCustomTopo(data: { id: number }) {
  return request({
    url: baseURL+'/api/run_custom_topo',
    method: 'get',
    params: {...data,'clientId':useLogStore().clientId}
  })
}

// 请求多机拓扑图数据
export function getTopoData() {
  return request({
    url: baseURL+'/api/multi_host',
    method: 'get'
  })
}
// 请求单个数图数据
export function getUuidTopo(data: { uuid: string }) {
  return request({
    url: baseURL+'/api/single_host_tree/' + data.uuid,
    method: 'get',
  })
}
// 请求host列表
export function getHostList() {
  return request({
    url: baseURL+'/api/agentlist',
    method: 'get',
  })
}
