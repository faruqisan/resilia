package kube

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetPods fuction return all pods name
func (e *Engine) GetPods() ([]string, error) {
	var (
		podNames []string
		err      error
	)

	pods, err := e.clientSet.CoreV1().Pods(e.namespace).List(metav1.ListOptions{})
	if err != nil {
		return podNames, err
	}

	for _, pod := range pods.Items {
		podNames = append(podNames, pod.Name)
	}

	return podNames, err
}
