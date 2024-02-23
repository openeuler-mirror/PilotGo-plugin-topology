import request from './request';

// 请求拓扑图数据
export function getTopoData() {
  return request({
    url: '/plugin/topology/api/multi_host',
    method: 'get'
  })
}
// 请求单个拓扑图数据
export function getSingleTopo(node: string) {
  return request({
    url: '/plugin/topology/api/single_host',
    method: 'get',
    params: node
  })
}
// 请求单个数图数据
export function getSingleTree(node: string) {
  return request({
    url: '/plugin/topology/api/single_host_tree',
    method: 'get',
    params: node
  })
}
// 请求host列表
export function getHostList() {
  return request({
    url: '/plugin/topology/api/agentlist',
    method: 'get'
  })
}
