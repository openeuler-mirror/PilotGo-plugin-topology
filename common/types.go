package common

// Metric defines a common metric interface
type Metric interface {
	// part of the Getter interface
	GetFieldInt64(field string) (int64, error)
	GetFieldKeys() []string

	Add(m Metric) Metric
	Sub(m Metric) Metric
	Split(cut int64) (Metric, Metric)
	GetStart() int64
	SetStart(start int64)
	GetLast() int64
	SetLast(last int64)
	IsZero() bool
}
