package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"atomgit.com/openeuler/PilotGo-plugin-topology/graffiti/rbac"
	auth "github.com/abbot/go-http-auth"
	"github.com/gorilla/mux"
	"github.com/skydive-project/dede/statics"
	"github.com/skydive-project/skydive/graffiti/api/rest"
	api "github.com/skydive-project/skydive/graffiti/api/server"
	"github.com/skydive-project/skydive/graffiti/api/types"
	shttp "github.com/skydive-project/skydive/graffiti/http"
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

func (w *WorkflowAPIHandler) executeWorkflow(resp http.ResponseWriter, r *auth.AuthenticatedRequest) {
	if !rbac.Enforce(r.Username, "workflow.call", "write") {
		http.Error(resp, http.StatusText(http.StatusUnauthorized), http.StatusMethodNotAllowed)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var wfCall types.WorkflowCall
	if err := decoder.Decode(&wfCall); err != nil {
		http.Error(resp, err.Error(), http.StatusBadRequest)
		return
	}

	vars := mux.Vars(&r.Request)

	resource, ok := w.Get(vars["ID"])
	if !ok {
		http.Error(resp, fmt.Sprintf("no workflow '%s' was found", vars["ID"]), http.StatusNotFound)
		return
	}
	workflow := resource.(*types.Workflow)

	ottoResult, err := w.runtime.ExecFunction(workflow.Source, wfCall.Params...)
	if err != nil {
		http.Error(resp, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := ottoResult.Export()
	if err != nil {
		http.Error(resp, err.Error(), http.StatusBadRequest)
		return
	}

	resp.Header().Set("Content-Type", "application/json; charset=UTF-8")
	resp.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(resp).Encode(result); err != nil {
		panic(err)
	}
}

// RegisterWorkflowAPI registers a new workflow api handler
func RegisterWorkflowAPI(apiServer *api.Server, authBackend shttp.AuthenticationBackend, runtime *js.Runtime) (*WorkflowAPIHandler, error) {
	workflowAPIHandler := &WorkflowAPIHandler{
		BasicAPIHandler: rest.BasicAPIHandler{
			ResourceHandler: &WorkflowResourceHandler{},
			EtcdClient:      apiServer.EtcdClient,
		},
		runtime: runtime,
	}
	apiServer.RegisterAPIHandler(workflowAPIHandler, authBackend)
	workflowAPIHandler.registerEndPoints(apiServer.HTTPServer, authBackend)
	return workflowAPIHandler, nil
}
