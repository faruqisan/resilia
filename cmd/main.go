package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/faruqisan/resilia/pkg/kube"
	"github.com/faruqisan/resilia/pumba"
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
		done = make(chan os.Signal, 1)
	)

	signal.Notify(done, os.Interrupt, syscall.SIGKILL, syscall.SIGINT, syscall.SIGTERM)

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
			}

		}
	}

	// spawn pumba!
	pumbaEngine := pumba.New(kubeEngine)

	// netem example
	// create netem worker to hit 1 redis
	// opt := pumbaEngine.NetEm(pumba.CommandNetEmLoss, pumba.NetEmOptions{
	// 	TCImage:     "gaiadocker/iproute2",
	// 	Duration:    "30s",
	// 	LossPercent: "90",
	// })

	opt := pumbaEngine.Pause(pumba.PauseOptions{
		Duration: "30s",
	})

	// get pods
	// wait for pods spawned
	var pods []string
	for {
		pods, err = kubeEngine.GetPods()
		if err != nil {
			log.Println(err)
		}
		log.Println("pods : ", pods)

		if len(pods) == 0 {
			log.Println("no pods found, waiting ...")
			time.Sleep(time.Second)
		} else {
			break
		}
	}

	// select 1st pod
	log.Println("targeting pods : ", pods[0])
	// to test file deployment
	// fData, err := ioutil.ReadFile(path.Join("files/examples/daemon_sets", "pumba.yaml"))

	pauseWorker := pumbaEngine.NewPumbaWorker(pods[0], "1m", pumba.CommandPause, opt)
	// netEmWorker, err := kubeEngine.LoadDaemonSetFromFile(fData)
	if err != nil {
		log.Fatal(err)
	}

	dName, err := pumbaEngine.RunWorker(pauseWorker)
	// dName, err := kubeEngine.CreateDaemonSet(netEmWorker)
	if err != nil {
		log.Println("fail to run pumba : ", err)
	} else {
		log.Printf("daemon set %s created", dName)
	}

	for {
		select {
		case <-done:
			var wg sync.WaitGroup

			wg.Add(3)
			go func() {
				defer wg.Done()
				terminateDeployments(kubeEngine)
			}()
			go func() {
				defer wg.Done()
				terminateServices(kubeEngine)
			}()
			go func() {
				defer wg.Done()
				terminateDaemonSets(kubeEngine)
			}()
			wg.Wait()
			return
		}
	}
}

func terminateDeployments(kubeEngine *kube.Engine) {
	createdDeployments, err := kubeEngine.GetDeployments()
	if err != nil {
		log.Println("err on delete deployment ", createdDeployments, err)
		return
	}
	for _, createdDeployment := range createdDeployments {
		err := kubeEngine.DeleteDeployment(createdDeployment)
		if err != nil {
			log.Println("err on delete deployment ", createdDeployment, err)
			continue
		}
		log.Printf("deployment %s deleted", createdDeployment)
	}
}

func terminateServices(kubeEngine *kube.Engine) {
	createdServices, err := kubeEngine.GetServices()
	if err != nil {
		log.Println("err on delete service ", createdServices, err)
		return
	}
	for _, createdService := range createdServices {
		err := kubeEngine.DeleteService(createdService)
		if err != nil {
			log.Println("err on delete service ", createdService, err)
			continue
		}
		log.Printf("service %s deleted", createdService)
	}
}

func terminateDaemonSets(kubeEngine *kube.Engine) {
	createdDaemonSets, err := kubeEngine.GetDaemonSets()
	if err != nil {
		log.Println("err on delete daemonset ", createdDaemonSets, err)
		return
	}
	for _, createdDaemonSet := range createdDaemonSets {
		err := kubeEngine.DeleteDaemonSet(createdDaemonSet)
		if err != nil {
			log.Println("err on delete daemonset ", createdDaemonSet, err)
			continue
		}
		log.Printf("daemonset %s deleted", createdDaemonSet)
	}
}
