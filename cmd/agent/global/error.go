/*
 * Copyright (c) KylinSoft  Co., Ltd. 2024.All rights reserved.
 * PilotGo-plugin-topology licensed under the Mulan Permissive Software License, Version 2. 
 * See LICENSE file for more details.
 * Author: Wangjunqi123 <wangjunqi@kylinos.cn>
 * Date: Fri Nov 8 09:13:05 2024 +0800
 */
package global

import "runtime"

func CallerInfo(err error) (string, int, string) {
	if err != nil {
		pro_c, filepath, line, ok := runtime.Caller(1)
		if ok {
			return filepath, line - 2, runtime.FuncForPC(pro_c).Name()
		}
	}
	return "", -1, ""
}
