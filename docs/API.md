# pilotgo-topo插件API接口

## 1. 标准SDK API接口
### 1.1 获取插件信息
- GET /plugin_manage/info
### 1.2 绑定pilotgo server
- PUT /plugin_manage/bind
### 1.3 event总线相关
- GET /plugin_manage/api/v1/extensions
- GET /plugin_manage/api/v1/gettags
- POST /plugin_manage/api/v1/event
- PUT /plugin_manage/api/v1/command_result

## 2. server端 API接口
### 2.1 获取agent列表
- GET /plugin/topology/api/agentlist
### 2.2 图数据时间戳列表
- GET /plugin/topology/api/timestamps
### 2.3 topo-agent心跳监听
- POST /plugin/topology/api/heartbeat
### 2.4 单机图数据
- GET /plugin/topology/api/single_host_tree/:uuid
### 2.5 多机网络拓扑图数据
- GET /plugin/topology/api/multi_host
### 2.6 创建自定义拓扑配置
- POST /plugin/topology/api/create_custom_topo
### 2.7 删除自定义拓扑配置
- DELETE /plugin/topology/api/delete_custom_topo?id=1
### 2.8 更新自定义拓扑配置
- PUT /plugin/topology/api/update_custom_topo?id=1
### 2.9 自定义拓扑配置列表
- GET /plugin/topology/api/custom_topo_list
### 2.10 调用自定义拓扑配置
- GET /plugin/topology/api/run_custom_topo?id=1

## 3. agent端 API接口
### 3.1 获取agent端采集数据
- GET /plugin/topology/api/rawdata