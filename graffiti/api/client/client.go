package client

// Operation describes a JSONPatch operation
type Operation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

// JSONPatch describes a JSON patch
type JSONPatch []Operation

// NewPatchOperation creates a new JSON patch operation
func NewPatchOperation(op, path string, value ...interface{}) Operation {
	patchOperation := Operation{
		Op:   op,
		Path: path,
	}
	if len(value) > 0 {
		patchOperation.Value = value[0]
	}
	return patchOperation
}
