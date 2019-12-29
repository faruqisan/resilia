package kube

import (
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	appsv1 "k8s.io/api/apps/v1"
)

// LoadDaemonSetFromFile function receive readed file in forms of byte
// and return the k8s daemonset object
func (e *Engine) LoadDaemonSetFromFile(file []byte) (*appsv1.DaemonSet, error) {

	var (
		daemonSet *appsv1.DaemonSet
		err       error
	)

	decode := scheme.Codecs.UniversalDeserializer().Decode

	obj, _, err := decode(file, nil, nil)
	if err != nil {
		return daemonSet, err
	}

	daemonSet = obj.(*appsv1.DaemonSet)

	return daemonSet, err
}

// IsDaemonSetExist function check wheter daemon set exist on cluster
func (e *Engine) IsDaemonSetExist(name string) (bool, error) {

	_, err := e.daemonSetsClient.Get(name, metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return false, err
	}

	return !errors.IsNotFound(err), nil
}

// GetDaemonSets function will return list of existing daemon set
func (e *Engine) GetDaemonSets() ([]string, error) {
	var (
		daemonSets []string
		err        error
	)

	ls, err := e.daemonSetsClient.List(metav1.ListOptions{})
	if err != nil {
		return daemonSets, err
	}

	for _, l := range ls.Items {
		daemonSets = append(daemonSets, l.Name)
	}

	return daemonSets, err
}

// CreateDaemonSet function will create a new daemon set on cluster
// returning created daemon set info (name)
func (e *Engine) CreateDaemonSet(daemonSet *appsv1.DaemonSet) (string, error) {
	result, err := e.daemonSetsClient.Create(daemonSet)
	if err != nil {
		return "", err
	}

	return result.GetObjectMeta().GetName(), nil
}

// DeleteDaemonSet function will remove daemon set from cluster
func (e *Engine) DeleteDaemonSet(name string) error {
	return e.daemonSetsClient.Delete(name, &metav1.DeleteOptions{})
}
