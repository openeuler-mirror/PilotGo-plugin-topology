import request from './request';
// 请求elk日志信息
export function getELKLogData(data:any) {
  return request({
    url: '/plugin/elk/api/log_timeaxis_data',
    method: 'post',
    data
  })
}