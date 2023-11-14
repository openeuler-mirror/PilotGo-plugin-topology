import axios from 'axios';
import { type AxiosInstance } from 'axios';

const request: AxiosInstance = axios.create({
  // baseURL: 'http://192.168.241.129:9991/plugin/api/', // 根据你的API地址进行配置
  baseURL: '', // 根据你的API地址进行配置
  timeout: 10000, // 设置超时时间
});

export default request;
