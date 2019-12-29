package kube

import (
	"os"
	"path"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	appstypev1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// helper to convert int to pointer of int32
func int32Ptr(i int32) *int32 { return &i }

// Option type used to customize kube Engine
type Option func(*Engine)

type (
	// Engine struct wrap k8s client
	// and act as function receiver
	Engine struct {
		useOutsideConfig  bool
		namespace         string // cluster namespace
		config            *rest.Config
		clientSet         *kubernetes.Clientset
		deploymentsClient appstypev1.DeploymentInterface
		servicesClient    corev1.ServiceInterface
		daemonSetsClient  appstypev1.DaemonSetInterface
	}
)

// WithNamespace function set namespace to kube engine
func WithNamespace(namespace string) Option {
	return func(e *Engine) {
		e.namespace = namespace
	}
}

// WithOutsideClusterConfig function set the flag on kube engine to
// get the config file from local kube config
// use this if you run this app outside cluster
func WithOutsideClusterConfig() Option {
	return func(e *Engine) {
		e.useOutsideConfig = true
	}
}

// New function return setuped engine
// by default k8s client will use InClusterConfig
// if this app didn't run on cluster
// use Option for non cluster
func New(options ...Option) (*Engine, error) {

	var (
		e    = new(Engine)
		conf *rest.Config
		err  error
	)

	for _, option := range options {
		option(e)
	}

	if e.namespace == "" {
		e.namespace = apiv1.NamespaceDefault
	}

	if e.useOutsideConfig {
		kubeConfigPath := path.Join(homeDir(), ".kube", "config")
		conf, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
		if err != nil {
			return e, err
		}
	} else {
		conf, err = rest.InClusterConfig()
		if err != nil {
			return e, err
		}
	}
	e.config = conf

	clientSet, err := kubernetes.NewForConfig(e.config)
	if err != nil {
		return e, err
	}
	e.clientSet = clientSet

	deploymentsClient := e.clientSet.AppsV1().Deployments(e.namespace)
	servicesClient := e.clientSet.CoreV1().Services(e.namespace)
	daemonSetsClient := e.clientSet.AppsV1().DaemonSets(e.namespace)

	e.deploymentsClient = deploymentsClient
	e.servicesClient = servicesClient
	e.daemonSetsClient = daemonSetsClient

	return e, nil
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
