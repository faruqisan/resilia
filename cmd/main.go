package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"

	"github.com/faruqisan/resilience_k8s/pkg/kube"
)

const (
	exampleDeploymentsFilesPath = "files/examples/deployments"
	exampleServiceFilesPath     = "files/examples/services"
)

func main() {

	// setup k8s
	kubeEngine, err := kube.New(kube.WithOutsideClusterConfig())
	if err != nil {
		log.Fatal(err)
	}

	// real all deployments
	// get file list from files folder
	files, err := ioutil.ReadDir(exampleDeploymentsFilesPath)
	if err != nil {
		log.Fatal(err)
	}

	var (
		createdDeployments []string
		createdServices    []string
		done               = make(chan os.Signal, 1)
	)

	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

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
				createdDeployments = append(createdDeployments, createdDeployment)

				log.Printf("deployment %s created", createdDeployment)
			}

			// -- get service
			// check if this deployments had service with same name
			servicePath := path.Join(exampleServiceFilesPath, file.Name())
			if _, err := os.Stat(servicePath); os.IsNotExist(err) {
				//skip to next loop
				log.Printf("file %s doesn't exist, err :%s \n", servicePath, err.Error())
				log.Printf("deployment : %s, doesn't have service, continue to next file\n", file.Name())
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
				createdServices = append(createdServices, createdService)
			}

		}
	}

	for {
		select {
		case <-done:
			log.Println("shutting down, cleaning up all deployments and services")
			for _, createdDeployment := range createdDeployments {
				err := kubeEngine.DeleteDeployment(createdDeployment)
				if err != nil {
					log.Println("err on delete deployment ", createdDeployment, err)
					continue
				}
				log.Printf("deployment %s deleted", createdDeployment)
			}

			for _, createdService := range createdServices {
				err := kubeEngine.DeleteService(createdService)
				if err != nil {
					log.Println("err on delete service ", createdService, err)
					continue
				}
				log.Printf("service %s deleted", createdService)
			}
			return
		}
	}
}
