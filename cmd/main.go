package main

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"github.com/faruqisan/resilience_k8s/pkg/kube"
)

const (
	exampleDeploymentsFilesPath = "files/examples/deployments"
	exampleServiceFilesPath     = "files/examples/services"
)

func main() {

	// setup k8s
	kubeEngine, err := kube.New("default")
	if err != nil {
		log.Fatal(err)
	}

	// real all deployments
	// get file list from files folder
	files, err := ioutil.ReadDir(exampleDeploymentsFilesPath)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		// process yaml only
		if strings.HasSuffix(file.Name(), ".yaml") {

			//read file
			// -- deployment 1st
			fData, err := ioutil.ReadFile(path.Join(exampleDeploymentsFilesPath, file.Name()))
			if err != nil {
				log.Fatal(err)
			}

			// create container from file data
			deployment, err := kubeEngine.LoadDeploymentFromFile(fData)
			if err != nil {
				log.Fatal(err)
			}

			deploymentExist, err := kubeEngine.IsDeploymentExist(deployment.Name)
			if err != nil {
				log.Println(err)
			}

			if !deploymentExist {
				createdDeployment, err := kubeEngine.CreateDeployment(deployment)
				if err != nil {
					log.Fatal(err)
				}

				log.Printf("deployment %s created", createdDeployment)
			}

			// -- get service
			// check if this deployments had service with same name
			servicePath := path.Join(exampleServiceFilesPath, file.Name())
			if _, err := os.Stat(servicePath); os.IsNotExist(err) {
				//skip to next loop
				log.Printf("deployment : %s, doesn't have service, continue to next file", file.Name())
				continue
			}

			fData, err = ioutil.ReadFile(servicePath)
			if err != nil {
				log.Fatal(err)
			}

			// create service from file data
			service, err := kubeEngine.LoadServiceFromFile(fData)
			if err != nil {
				log.Fatal(err)
			}

			serviceExist, err := kubeEngine.IsServiceExist(service.Name)
			if err != nil {
				log.Println(err)
			}

			if !serviceExist {
				createdService, err := kubeEngine.CreateService(service)
				if err != nil {
					log.Fatal(err)
				}

				log.Printf("service %s created", createdService)
			}

		}
	}

	for {
	}
}
