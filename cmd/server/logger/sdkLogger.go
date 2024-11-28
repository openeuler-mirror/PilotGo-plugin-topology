/*
 * Copyright (c) KylinSoft  Co., Ltd. 2024.All rights reserved.
 * PilotGo-plugin-topology licensed under the Mulan Permissive Software License, Version 2. 
 * See LICENSE file for more details.
 * Author: Wangjunqi123 <wangjunqi@kylinos.cn>
 * Date: Mon Nov 4 14:30:13 2024 +0800
 */
package logger

import (
	"fmt"
	"os"

	"gitee.com/openeuler/PilotGo-plugin-topology/cmd/server/conf"
	"gitee.com/openeuler/PilotGo/sdk/logger"
)

func InitLogger() {
	err := logger.Init(conf.Global_Config.Logopts)
	if err != nil {
		fmt.Printf("logger module init failed: %s\n", err.Error())
		os.Exit(1)
	}
}
