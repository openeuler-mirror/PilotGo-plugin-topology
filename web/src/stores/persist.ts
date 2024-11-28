/* 
 * Copyright (c) KylinSoft  Co., Ltd. 2024.All rights reserved.
 * PilotGo-plugin-topology licensed under the Mulan Permissive Software License, Version 2. 
 * See LICENSE file for more details.
 * Author: zhaozhenfang <zhaozhenfang@kylinos.cn>
 * Date: Fri Jul 5 15:34:21 2024 +0800
 */
// store数据持久化
import { createPinia } from 'pinia'
import piniaPersisted from 'pinia-plugin-persistedstate'
const pinia = createPinia();
pinia.use(piniaPersisted);

export default pinia;