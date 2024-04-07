package processor

import "strings"

func ProcessMatching(exename, cmdline, component string) bool {
	component_lower := strings.ToLower(component)
	cmdline_lower := strings.ToLower(cmdline)
	cmdline_lower_arr := strings.Split(cmdline_lower, " ")

	switch exename {
	case "java":
		mainclass_match := false
		match_count := 0

		for i := 1; i < len(cmdline_lower_arr); i++ {
			if !strings.HasPrefix(cmdline_lower_arr[i], "-") && !strings.HasPrefix(cmdline_lower_arr[i], "/") {
				if strings.Contains(cmdline_lower_arr[i], component_lower) {
					mainclass_match = true
				}
			}

			if strings.Contains(cmdline_lower_arr[i], component_lower) {
				match_count = match_count + 1
			}

			if mainclass_match && match_count >= 1 {
				return true
			}
		}
	case "python", "python3", "python2":
		for i := 1; i < len(cmdline_lower_arr); i++ {
			if strings.HasSuffix(cmdline_lower_arr[i], ".py") && cmdline_lower_arr[i] == component_lower {
				return true
			}
		}
	case "ruby":
		for i := 1; i < len(cmdline_lower_arr); i++ {
			if strings.HasSuffix(cmdline_lower_arr[i], ".rb") && cmdline_lower_arr[i] == component_lower {
				return true

			}
		}
	case "node":
		for i := 1; i < len(cmdline_lower_arr); i++ {
			if strings.HasSuffix(cmdline_lower_arr[i], ".js") || strings.HasSuffix(cmdline_lower_arr[i], ".ts") && cmdline_lower_arr[i] == component_lower {
				return true
			}
		}
	case "perl":
		for i := 1; i < len(cmdline_lower_arr); i++ {
			if strings.HasSuffix(cmdline_lower_arr[i], ".pl") && cmdline_lower_arr[i] == component_lower {
				return true
			}
		}
	default:
		if strings.Contains(exename, component_lower) {
			return true
		}
	}

	return false
}
