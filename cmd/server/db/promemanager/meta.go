package promemanager

type Target struct {
	Instance string            `json:"instance"`
	Job      string            `json:"job"`
	Metrics  map[string]string `json:"metrics"`
}