// Package services services hold all business logic for preset suites domain
package services

import (
	"github.com/faruqisan/resilia/pkg/pumba"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

type (

	// KubeEngine interface define kube engine required contract
	// this is helping us to mock kube package
	KubeEngine interface {
		LoadDeploymentFromFile(file []byte) (*appsv1.Deployment, error)
		IsDeploymentExist(name string) (bool, error)
		CreateDeployment(deployment *appsv1.Deployment) (string, error)
		GetDeployments() ([]string, error)
		DeleteDeployment(name string) error
		LoadDaemonSetFromFile(file []byte) (*appsv1.DaemonSet, error)
		IsDaemonSetExist(name string) (bool, error)
		GetDaemonSets() ([]string, error)
		CreateDaemonSet(daemonSet *appsv1.DaemonSet) (string, error)
		DeleteDaemonSet(name string) error
		LoadServiceFromFile(file []byte) (*corev1.Service, error)
		IsServiceExist(name string) (bool, error)
		GetServices() ([]string, error)
		CreateService(service *corev1.Service) (string, error)
		DeleteService(name string) error
		GetPods() ([]string, error)
	}

	// PumbaEngine interface define pumba engine required contract
	// this is helping us to mock pumba package
	PumbaEngine interface {
		NetEm(netEmCommand pumba.NetEmCommands, options pumba.NetEmOptions) pumba.WorkerOptions
		Pause(options pumba.PauseOptions) pumba.WorkerOptions
		NewPumbaWorker(target string, interval string, mode pumba.WorkerCommandMode, options ...pumba.WorkerOptions) pumba.Worker
		RunWorker(worker pumba.Worker) (string, error)
	}

	// KubeKind type define k8s resource kind, eg : deployment, resources or daemon set
	KubeKind string

	// FileResource struct hold test suites k8s resource data
	// like deployment, services, daemon sets and other
	FileResource struct {
		ID      string   `json:"id"`
		SuiteID string   `json:"suite_id"`
		Name    string   `json:"name"`
		Kind    KubeKind `json:"king"`
		Value   string   `json:"value"` // k8s deployment.yaml file that parsed into json and stored in database as string
	}

	// Model struct define test suites
	// never access property of this struct directly
	Model struct {
		ID               string         `json:"id"`
		Name             string         `json:"name"`
		Resources        []FileResource `json:"resources,omitempty"`
		pumbaWorkers     []pumba.Worker
		CreatedResources map[KubeKind][]string `json:"created_resources,omitempty"`
	}

	// Service struct hold all requirement for suites services
	Service struct {
		kubeEngine  KubeEngine
		pumbaEngine PumbaEngine
	}
)

const (
	// KindDeployment is k8s kind for deployment
	KindDeployment KubeKind = "deployment"
	// KindService is k8s kind for service
	KindService KubeKind = "service"
	// KindDaemonSet is k8s kind for general daemon set
	// never put pumba dameon set as this kind
	KindDaemonSet KubeKind = "daemonset"
	// KindPumbaDaemonSet is k8s kind for pumba daemon set
	KindPumbaDaemonSet KubeKind = "daemonset"
)

// New function return new service object with setuped requirement
func New(kubeEngine KubeEngine, pumbaEngine PumbaEngine) *Service {
	return &Service{
		kubeEngine:  kubeEngine,
		pumbaEngine: pumbaEngine,
	}
}

// Create function create a new suite model and store it to database
func (s *Service) Create(id string, name string) *Model {
	return &Model{
		ID:               id,
		Name:             name,
		CreatedResources: make(map[KubeKind][]string),
	}
}

// AddPumbaWorker function add pumba worker into suite
func (m *Model) AddPumbaWorker(worker pumba.Worker) {
	m.pumbaWorkers = append(m.pumbaWorkers, worker)
}

// RunSuiteFileResources function run given suite model's only
// file resource, you can add pumba worker later
func (s *Service) RunSuiteFileResources(suite *Model) error {

	// load all resouce and apply it
	for _, resource := range suite.Resources {
		if resource.Kind != KindDaemonSet {
			name, err := s.applyResourceValue(resource)
			if err != nil {
				return err
			}

			suite.CreatedResources[resource.Kind] = append(suite.CreatedResources[resource.Kind], name)
		}
	}

	return nil
}

// RunSuitePumbaWorkers function run only suite's pumba worker
func (s *Service) RunSuitePumbaWorkers(suite *Model) error {
	// load all pumba worker
	for _, worker := range suite.pumbaWorkers {
		name, err := s.pumbaEngine.RunWorker(worker)
		if err != nil {
			return err
		}
		suite.CreatedResources[KindPumbaDaemonSet] = append(suite.CreatedResources[KindPumbaDaemonSet], name)
	}
	return nil
}

// StopSuites function delete all created resources during suite test
// this will also clean the created resources
func (s *Service) StopSuites(suite *Model) error {
	for kind, createdResources := range suite.CreatedResources {
		switch kind {
		case KindDeployment:
			if err := s.terminateDeployments(createdResources);err != nil {
				return err
			}
		case KindService:
			if err := s.terminateServices(createdResources); err != nil {
				return err
			}
		case KindDaemonSet:
			if err := s.terminateDaemonSets(createdResources); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Service) applyResourceValue(resource FileResource) (string, error) {

	jsonData := []byte(resource.Value)

	switch resource.Kind {
	case KindDeployment:
		return s.applyDeployment(jsonData)
	case KindService:
		return s.applyService(jsonData)
	case KindDaemonSet:
		return s.applyDaemonSet(jsonData)
	}

	return "", nil
}

func (s *Service) applyDeployment(value []byte) (string, error) {

	dep, err := s.kubeEngine.LoadDeploymentFromFile(value)
	if err != nil {
		return "", err
	}

	return s.kubeEngine.CreateDeployment(dep)

}

func (s *Service) applyService(value []byte) (string, error) {

	svc, err := s.kubeEngine.LoadServiceFromFile(value)
	if err != nil {
		return "", err
	}

	return s.kubeEngine.CreateService(svc)

}

func (s *Service) applyDaemonSet(value []byte) (string, error) {

	ds, err := s.kubeEngine.LoadDaemonSetFromFile(value)
	if err != nil {
		return "", err
	}

	return s.kubeEngine.CreateDaemonSet(ds)

}

func (s *Service) terminateDeployments(createdResources []string) error {
	for _, createdDeployment := range createdResources {
		err := s.kubeEngine.DeleteDeployment(createdDeployment)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) terminateServices(createdResources []string) error {
	for _, createdService := range createdResources {
		err := s.kubeEngine.DeleteService(createdService)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) terminateDaemonSets(createdResources []string) error {
	for _, createdDaemonSet := range createdResources {
		err := s.kubeEngine.DeleteDaemonSet(createdDaemonSet)
		if err != nil {
			return err
		}
	}
	return nil
}
