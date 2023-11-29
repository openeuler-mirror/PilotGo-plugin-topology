# pilotgo-topo插件server端API接口
## 标准SDK API接口
### 获取插件信息
- GET /plugin_manage/info
### 绑定pilotgo server
- PUT /plugin_manage/bind
### event总线相关
- GET /plugin_manage/api/v1/extensions
- GET /plugin_manage/api/v1/gettags
- POST /plugin_manage/api/v1/event
- PUT /plugin_manage/api/v1/command_result