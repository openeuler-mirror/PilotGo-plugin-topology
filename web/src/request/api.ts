import request from './request';

export const topo = {
  async multi_host_topo() {
    try {
      const response = await request.get('/plugin/topology/api/multi_host');
      return response.data;
    } catch (error) {
      throw error;
    }
  },
  async single_host_topo(node:string) {
    try {
      const response = await request.get('/plugin/topology/api/single_host/'+node);
      return response.data;
    } catch (error) {
      throw error;
    }
  },

  async single_host_tree(node:string) {
    try {
      const response = await request.get('/plugin/topology/api/single_host_tree/'+node);
      return response.data;
    } catch (error) {
      throw error;
    }
  },

  async host_list() {
    try {
      const response = await request.get('/plugin/topology/api/agentlist');
      return response.data;
    } catch (error) {
      throw error;
    }
  },
  
  // 添加其他API请求方法
};
