import request from './request';

// 请求批次信息
export function getBatchList() {
  return request({
    url: '/api/batch_list',
    method: 'get',
  })
}

// 请求批次详情
export function getBatchDetail(data: { batchId: number }) {
  return request({
    url: '/api/batch_uuid',
    method: 'get',
    params: data
  })
}

// 请求拓扑配置列表
export function getConfList(data: {}) {
  return request({
    url: '/api/custom_topo_list',
    method: 'get',
    params: data
  })
}

// 新增拓扑配置列表
export function addConfList(data: {}) {
  return request({
    url: '/api/create_custom_topo',
    method: 'POST',
    data
  })
}

// 更新拓扑配置列表
export function updateConfList(data: {}) {
  return request({
    url: '/api/update_custom_topo',
    method: 'put',
    data
  })
}


// 删除拓扑配置
export function delConfig(data: { id: number }) {
  return request({
    url: '/api/delete_custom_topo',
    method: 'delete',
    data: data
  })
}

// 请求某一个拓扑图数据
export function getCustomTopo(data: { id: number }) {
  return request({
    url: '/api/run_custom_topo',
    method: 'get',
    params: data
  })
}

// 请求多机拓扑图数据
export function getTopoData() {
  return request({
    url: '/api/multi_host',
    method: 'get'
  })
}
// 请求单个数图数据
export function getUuidTopo(data: { uuid: string }) {
  return request({
    url: '/api/single_host_tree/' + data.uuid,
    method: 'get',
  })
}
// 请求host列表
export function getHostList() {
  return request({
    url: '/api/agentlist',
    method: 'get',
  })
}
