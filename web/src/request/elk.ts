import request from './request';
// 请求elk日志信息
export function getLogData(data:{}) {
  return request({
    url: '/plugin/elk/api/search',
    method: 'post',
    data
  })
}