package meta

import "time"

type TopoConfiguration struct {
	ID          uint      `gorm:"primary_key;AUTO_INCREMENT" json:"id"`
	Name        string    `gorm:"not null;type:varchar(200)" json:"conf_name"`
	Version     string    `gorm:"not null;type:varchar(20)" json:"conf_version"`
	Description string    `gorm:"type:text" json:"description"`
	CreatedAt  time.Time `gorm:"not null" json:"conf_time"`
	Preserve    uint      `gorm:"not null" json:"preserve"`
	Machines    string    `gorm:"not null;type:text" json:"machines"`
	NodeRules   string    `gorm:"type:text" json:"node_rules"`
	TagRules    string    `gorm:"type:text" json:"tag_rules"`
}
