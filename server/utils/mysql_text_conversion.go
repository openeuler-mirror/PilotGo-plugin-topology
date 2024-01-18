package utils

import (
	"encoding/json"

	"gitee.com/openeuler/PilotGo-plugin-topology-server/meta"
	"github.com/pkg/errors"
)

func TextConversionFromMysql(tc *meta.Topo_configuration) (*[]string, *[][]meta.Filter_rule, *[]meta.Tag_rule, error) {
	machines := make([]string, 0)
	node_rules := make([][]meta.Filter_rule, 0)
	tag_rules := make([]meta.Tag_rule, 0)

	// machines数据处理
	ptr, ok := (tc.Machines).(*interface{})
	if !ok {
		return nil, nil, nil, errors.New("machines type assert error **warn**2")
	}

	ASCII_bytes, ok := (*ptr).([]uint8)
	if !ok {
		return nil, nil, nil, errors.New("machines type assert error **warn**2")
	}

	err := json.Unmarshal(ASCII_bytes, &machines)
	if err != nil {
		return nil, nil, nil, errors.Errorf("machines json unmarshal error: %s **warn**2", err.Error())
	}

	// node_rules数据处理
	ptr, ok = (tc.NodeRules).(*interface{})
	if !ok {
		return nil, nil, nil, errors.New("node_rules type assert error **warn**2")
	}

	ASCII_bytes, ok = (*ptr).([]uint8)
	if !ok {
		return nil, nil, nil, errors.Errorf("node_rules type assert error **warn**2")
	}

	err = json.Unmarshal(ASCII_bytes, &node_rules)
	if err != nil {
		return nil, nil, nil, errors.Errorf("node_rules json unmarshal error: %s **warn**2", err.Error())
	}

	// tag_rules数据处理
	ptr, ok = (tc.TagRules).(*interface{})
	if !ok {
		return nil, nil, nil, errors.New("tag_rules type assert error **warn**2")
	}

	ASCII_bytes, ok = (*ptr).([]uint8)
	if !ok {
		return nil, nil, nil, errors.Errorf("tag_rules type assert error **warn**2")
	}

	err = json.Unmarshal(ASCII_bytes, &tag_rules)
	if err != nil {
		return nil, nil, nil, errors.Errorf("tag_rules json unmarshal error: %s **warn**2", err.Error())
	}

	return &machines, &node_rules, &tag_rules, nil
}
