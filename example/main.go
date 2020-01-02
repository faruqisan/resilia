package main

import (
	"flag"
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
	"github.com/faruqisan/resilia/pkg/pumba"
)

const (
	exampleDeploymentsFilesPath = "files/deployments"
	exampleServiceFilesPath     = "files/services"
	exampleDaemonSetsPath       = "files/daemon_sets"
)

var (
	inCluster bool
)

func main() {

	flag.BoolVar(&inCluster, "in_cluster", false, "flag if this app run inside k8s cluster")
	flag.Parse()

	var (
		kubeEngine *kube.Engine
		err        error
	)

	// setup k8s
	switch inCluster {
	case true:
		kubeEngine, err = kube.New()
		if err != nil {
			log.Fatal(err)
		}
	case false:
		kubeEngine, err = kube.New(kube.WithOutsideClusterConfig())
		if err != nil {
			log.Fatal(err)
		}
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

	// get pods
	// wait for pods spawned
	var (
		pods     []string
		redisPod string
	)
	for {
		pods, err = kubeEngine.GetPods()
		if err != nil {
			log.Println(err)
		}

		if len(pods) == 0 {
			log.Println("no pods found, waiting ...")
			time.Sleep(time.Second)
			continue
		}
		// target redis pods
		for _, pod := range pods {
			// check if pod name contain redis
			if strings.Contains(pod, "redis") {
				redisPod = pod
				break
			}
		}

		if redisPod == "" {
			continue
		}

		break
	}

	log.Println("targeting pods : ", redisPod)

	// spawn pumba!
	pumbaEngine := pumba.New(kubeEngine)
	opt := pumbaEngine.Pause(pumba.PauseOptions{
		Duration: "10s",
	})

	pauseWorker := pumbaEngine.NewPumbaWorker(pods[0], "20s", pumba.CommandPause, opt)
	if err != nil {
		log.Fatal(err)
	}

	dName, err := pumbaEngine.RunWorker(pauseWorker)
	if err != nil {
		log.Fatal("fail to run pumba : ", err)
	}
	log.Printf("daemon set %s created", dName)

	for {
		select {
		case <-done:
			cleanup(kubeEngine)
			return
		}
	}
}

func cleanup(kubeEngine *kube.Engine) {
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
