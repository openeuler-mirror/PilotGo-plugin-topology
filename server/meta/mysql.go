package meta

const (
	FILTER_TYPE_HOST     = "host"
	FILTER_TYPE_PROCESS  = "process"
	FILTER_TYPE_RESOURCE = "resource"
	FILTER_TYPE_TAG      = "tag"
)

type Topo_configuration struct {
	ID          uint            `gorm:"primary_key;AUTO_INCREMENT" json:"id"`
	Name        string          `gorm:"not null;type:varchar(200)" json:"conf_name"`
	Version     string          `gorm:"not null;type:varchar(20)" json:"conf_version"`
	Description string          `gorm:"type:text" json:"description"`
	CreatedAt   string          `gorm:"not null" json:"create_time"`
	UpdatedAt   string          `gorm:"not null" json:"update_time"`
	Preserve    uint            `gorm:"not null" json:"preserve"`
	Machines    []string        `gorm:"not null;type:text" json:"machines"`
	NodeRules   [][]Filter_rule `gorm:"type:text" json:"node_rules"`
	TagRules    []Tag_rule      `gorm:"type:text" json:"tag_rules"`
}

type Topo_configuration_DB struct {
	ID          uint   `gorm:"primary_key;AUTO_INCREMENT" json:"id"`
	Name        string `gorm:"not null;type:varchar(200)" json:"conf_name"`
	Version     string `gorm:"not null;type:varchar(20)" json:"conf_version"`
	Description string `gorm:"type:text" json:"description"`
	CreatedAt   string `gorm:"not null" json:"create_time"`
	UpdatedAt   string `gorm:"not null" json:"update_time"`
	Preserve    uint   `gorm:"not null" json:"preserve"`
	Machines    string `gorm:"not null;type:text" json:"machines"`
	NodeRules   string `gorm:"type:text" json:"node_rules"`
	TagRules    string `gorm:"type:text" json:"tag_rules"`
}

type Filter_rule struct {
	Rule_type      string            `json:"rule_type"`
	Rule_condition map[string]string `json:"rule_condition"`
}

type Tag_rule struct {
	Tag_name string          `json:"tag_name"`
	Target   string          `json:"target"`
	Rules    [][]Filter_rule `json:"rules"`
}
