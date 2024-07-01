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