package kube

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

// LoadServiceFromFile function receive readed file in forms of byte
// and return the k8s services object
func (e *Engine) LoadServiceFromFile(file []byte) (*corev1.Service, error) {

	var (
		s   *corev1.Service
		err error
	)

	decode := scheme.Codecs.UniversalDeserializer().Decode

	obj, _, err := decode(file, nil, nil)
	if err != nil {
		return s, err
	}

	s = obj.(*corev1.Service)

	return s, err
}

// IsServiceExist function check wheter service exist on cluster
func (e *Engine) IsServiceExist(name string) (bool, error) {

	_, err := e.servicesClient.Get(name, metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return false, err
	}

	return !errors.IsNotFound(err), nil
}

// GetServices function will return list of existing services
func (e *Engine) GetServices() ([]string, error) {
	var (
		serviceNames []string
		err          error
	)

	ls, err := e.servicesClient.List(metav1.ListOptions{})
	if err != nil {
		return serviceNames, err
	}

	for _, l := range ls.Items {
		serviceNames = append(serviceNames, l.Name)
	}

	return serviceNames, err
}

// CreateService function will create a new service on cluster
// returning created service info (name)
func (e *Engine) CreateService(service *corev1.Service) (string, error) {
	result, err := e.servicesClient.Create(service)
	if err != nil {
		return "", err
	}

	return result.GetObjectMeta().GetName(), nil
}

// DeleteService function will remove service from cluster
func (e *Engine) DeleteService(name string) error {
	return e.servicesClient.Delete(name, &metav1.DeleteOptions{})
}