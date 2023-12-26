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

## 2. server端API接口
### 2.1 获取agent列表
- GET /plugin/topology/api/agentlist
### 2.2 单机图数据
- GET /plugin/topology/api/single_host_tree/:uuid
### 2.3 多机网络拓扑图数据
- GET /plugin/topology/api/multi_host
### 2.4 topo-agent心跳监听
- POST /heartbeat

## 3. agent端API接口
### 3.1 获取agent端采集数据
- GET /plugin/topology/api/rawdata