package utils

import (
	"gitee.com/openeuler/PilotGo-plugin-topology-server/db/mysqlmanager"
	"gitee.com/openeuler/PilotGo-plugin-topology-server/graph"
	"github.com/pkg/errors"
)

func TagInjection(n *graph.Node, tags []mysqlmanager.Tag_rule) error {
	for _, tagrule := range tags {
		for _, rules := range tagrule.Rules {
			// 判断是否为同一台机器
			uuid := ""
			for _, condition := range rules {
				if condition.Rule_type == mysqlmanager.FILTER_TYPE_HOST {
					_uuid, ok := condition.Rule_condition["uuid"]
					if !ok {
						return errors.Errorf("there is no uuid field in tag rule_condition: %+v **errstack**2", condition.Rule_condition)
					}
					uuid = _uuid.(string)
					break
				}
			}
			if uuid != n.UUID {
				continue
			}

			// 为host节点添加标签
			if n.Type == "host" && len(rules) == 1 {
				n.Tags = append(n.Tags, tagrule.Tag_name)
				break
			}

			for _, condition := range rules {
				switch condition.Rule_type {
				case mysqlmanager.FILTER_TYPE_HOST:
					continue
				case mysqlmanager.FILTER_TYPE_PROCESS:
					if _name, ok := condition.Rule_condition["name"]; !ok {
						return errors.Errorf("there is no name field in tag rule_condition: %+v **errstack**2", condition.Rule_condition)
					} else if _name.(string) == n.Name {
						n.Tags = append(n.Tags, tagrule.Tag_name)
					}
				case mysqlmanager.FILTER_TYPE_TAG:
					if _tag, ok := condition.Rule_condition["tag_name"]; !ok {
						return errors.Errorf("there is no tag_name field in tag rule_condition: %+v **errstack**2", condition.Rule_condition)
					} else if _tag.(string) == n.Name {
						n.Tags = append(n.Tags, tagrule.Tag_name)
					}
				case mysqlmanager.FILTER_TYPE_RESOURCE:
					// TODO: 暂时不区分disk cpu nc等资源节点
					n.Tags = append(n.Tags, tagrule.Tag_name)
				}
			}
		}
	}

	return nil
}
