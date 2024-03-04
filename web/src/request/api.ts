import request from './request';

// 请求批次信息
export function getBatchList() {
  return request({
    url: '/plugin/topology/api/batch_list',
    method: 'get',
  })
}

// 请求批次详情
export function getBatchDetail(data: { batchId: number }) {
  return request({
    url: '/plugin/topology/api/batch_uuid',
    method: 'get',
    params: data
  })
}

// 请求拓扑配置列表
export function getConfList(data: {}) {
  return request({
    url: '/plugin/topology/api/custom_topo_list',
    method: 'get',
    params: data
  })
}

// 删除拓扑配置
export function delConfig(data: { id: number }) {
  return request({
    url: '/plugin/topology/api/delete_custom_topo',
    method: 'delete',
    data: data
  })
}

// 请求某一个拓扑图数据
export function getCustomTopo(data: { id: number }) {
  return request({
    url: '/plugin/topology/api/run_custom_topo',
    method: 'get',
    params: data
  })
}

// 请求多机拓扑图数据
export function getTopoData() {
  return request({
    url: '/plugin/topology/api/multi_host',
    method: 'get'
  })
}
// 请求单个数图数据
export function getUuidTopo(data: { uuid: string }) {
  return request({
    url: '/plugin/topology/api/single_host_tree/' + data.uuid,
    method: 'get',
  })
}
// 请求host列表
export function getHostList() {
  return request({
    url: '/plugin/topology/api/agentlist',
    method: 'get',
  })
}
