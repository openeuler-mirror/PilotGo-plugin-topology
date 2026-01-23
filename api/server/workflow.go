package server

import (
	"fmt"

	"github.com/skydive-project/dede/statics"
	"github.com/skydive-project/skydive/graffiti/api/rest"
	"github.com/skydive-project/skydive/graffiti/api/types"
	"github.com/skydive-project/skydive/graffiti/js"
	"github.com/skydive-project/skydive/graffiti/logging"
	"gopkg.in/yaml.v2"
)

const workflowAssetDir = "statics/workflows"

// WorkflowResourceHandler describes a workflow resource handler
type WorkflowResourceHandler struct {
}

// WorkflowAPIHandler based on BasicAPIHandler
type WorkflowAPIHandler struct {
	rest.BasicAPIHandler
	runtime *js.Runtime
}

// New creates a new workflow resource
func (w *WorkflowResourceHandler) New() rest.Resource {
	return &types.Workflow{}
}

// Name return "workflow"
func (w *WorkflowResourceHandler) Name() string {
	return "workflow"
}

// Create tests whether the resource is a duplicate or is unique
func (w *WorkflowAPIHandler) Create(r rest.Resource, opts *rest.CreateOptions) error {
	workflow := r.(*types.Workflow)

	for _, resource := range w.Index() {
		w := resource.(*types.Workflow)
		if w.Name == workflow.Name {
			return fmt.Errorf("Duplicate workflow, name=%s", w.Name)
		}
	}

	return w.BasicAPIHandler.Create(workflow, opts)
}

func (w *WorkflowAPIHandler) loadWorkflowAsset(name string) (*types.Workflow, error) {
	yml, err := statics.Asset(name)
	if err != nil {
		return nil, err
	}

	var workflow types.Workflow
	if err := yaml.Unmarshal([]byte(yml), &workflow); err != nil {
		return nil, err
	}

	return &workflow, nil
}

// Get retrieves a workflow based on its id
func (w *WorkflowAPIHandler) Get(id string) (rest.Resource, bool) {
	workflows := w.Index()
	workflow, found := workflows[id]
	if !found {
		return nil, false
	}
	return workflow.(*types.Workflow), true
}

// Index returns a map of workflows indexed by id
func (w *WorkflowAPIHandler) Index() map[string]rest.Resource {
	resources := w.BasicAPIHandler.Index()
	assets, err := statics.AssetDir(workflowAssetDir)
	if err == nil {
		for _, asset := range assets {
			workflow, err := w.loadWorkflowAsset(workflowAssetDir + "/" + asset)
			if err != nil {
				logging.GetLogger().Errorf("Failed to load worklow asset %s: %s", asset, err)
				continue
			}
			resources[workflow.GetID()] = workflow
		}
	}
	return resources
}
