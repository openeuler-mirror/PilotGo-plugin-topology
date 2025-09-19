package assets

// Assets is the interface by an asset provider
type Assets interface {
	Asset(name string) ([]byte, error)
}
