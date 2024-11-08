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
