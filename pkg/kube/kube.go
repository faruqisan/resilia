package kube

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	appstypev1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

// helper to convert int to pointer of int32
func int32Ptr(i int32) *int32 { return &i }

type (
	// Engine struct wrap k8s client
	// and act as function receiver
	Engine struct {
		namespace         string // cluster namespace
		config            *rest.Config
		clientSet         *kubernetes.Clientset
		deploymentsClient appstypev1.DeploymentInterface
		servicesClient    corev1.ServiceInterface
	}
)

// New function return setuped engine
func New(namespace string) (*Engine, error) {

	var (
		e   = new(Engine)
		err error
	)

	config, err := rest.InClusterConfig()
	if err != nil {
		return e, err
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return e, err
	}

	e.namespace = namespace
	if e.namespace == "" {
		e.namespace = apiv1.NamespaceDefault
	}

	deploymentsClient := clientSet.AppsV1().Deployments(e.namespace)
	servicesClient := clientSet.CoreV1().Services(e.namespace)

	e.config = config
	e.clientSet = clientSet
	e.deploymentsClient = deploymentsClient
	e.servicesClient = servicesClient

	return e, nil
}
