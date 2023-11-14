import request from './request'
// 获取指标列表
export function getPromRules() {
  return request({
    url: '/plugin/prometheus/api/v1/targets',
    method: 'get',
  })
}

// 获取prome某一时间点的数据
export function getPromeCurrent(data: object) {
  return request({
    url: '/plugin/prometheus/api/v1/query',
    method: 'get',
    params: data
  })
}

// 获取prome某一时间段的数据
export function getPromeRange(data: object) {
  return request({
    url: '/plugin/prometheus/api/v1/query_range',
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