package kube

import (
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

// LoadDeploymentFromFile function receive readed file in forms of byte
// and return the k8s deployment object
func (e *Engine) LoadDeploymentFromFile(file []byte) (*appsv1.Deployment, error) {

	var (
		deployment *appsv1.Deployment
		err        error
	)

	decode := scheme.Codecs.UniversalDeserializer().Decode

	obj, _, err := decode(file, nil, nil)
	if err != nil {
		return deployment, err
	}

	deployment = obj.(*appsv1.Deployment)

	return deployment, err
}

// GetDeployments function will return list of existing deployments
func (e *Engine) GetDeployments() ([]string, error) {
	var (
		deploymentNames []string
		err             error
	)

	ls, err := e.deploymentsClient.List(metav1.ListOptions{})
	if err != nil {
		return deploymentNames, err
	}

	for _, l := range ls.Items {
		deploymentNames = append(deploymentNames, l.Name)
	}

	return deploymentNames, err
}

// IsDeploymentExist function check wheter deployment exist on cluster
func (e *Engine) IsDeploymentExist(name string) (bool, error) {

	_, err := e.deploymentsClient.Get(name, metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return false, err
	}

	return !errors.IsNotFound(err), nil
}

// CreateDeployment function will create a new deployment on cluster
// returning created deployment info (name)
func (e *Engine) CreateDeployment(deployment *appsv1.Deployment) (string, error) {
	result, err := e.deploymentsClient.Create(deployment)
	if err != nil {
		return "", err
	}

	return result.GetObjectMeta().GetName(), nil
}

// DeleteDeployment function will remove deployment from cluster
func (e *Engine) DeleteDeployment(name string) error {
	return e.deploymentsClient.Delete(name, &metav1.DeleteOptions{})
}
