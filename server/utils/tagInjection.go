package utils

import "gitee.com/openeuler/PilotGo-plugin-topology-server/meta"

func TagInjection(n *meta.Node, tags []meta.Tag_rule) *meta.Node {
	for _, tagrule := range tags {
		for _, rules := range tagrule.Rules {
			// 判断是否为同一台机器
			uuid := ""
			for _, condition := range rules {
				if condition.Rule_type == meta.FILTER_TYPE_HOST {
					uuid = condition.Rule_condition["uuid"]
					break
				}
			}
			if uuid != n.UUID {
				continue
			}

			// 为host节点添加标签
			if len(rules) == 1 {
				n.Tags = append(n.Tags, tagrule.Tag_name)
				break
			}

			for _, condition := range rules {
				switch condition.Rule_type {
				case meta.FILTER_TYPE_HOST:
					continue
				case meta.FILTER_TYPE_PROCESS:
					if condition.Rule_condition["name"] == n.Name {
						n.Tags = append(n.Tags, tagrule.Tag_name)
					}
				case meta.FILTER_TYPE_TAG:
					if condition.Rule_condition["tag_name"] == n.Name {
						n.Tags = append(n.Tags, tagrule.Tag_name)
					}
				case meta.FILTER_TYPE_RESOURCE:
					// TODO: 暂时不区分disk cpu nc等资源节点
					n.Tags = append(n.Tags, tagrule.Tag_name)
				}
			}
		}
	}

	return n
}
