package main

import (
	"flag"
	"log"
	"time"

	httpserver "github.com/faruqisan/resilia/engine/servers/http"
	"github.com/faruqisan/resilia/engine/suites/services"
	"github.com/faruqisan/resilia/pkg/kube"
	"github.com/faruqisan/resilia/pkg/pumba"
)

var (
	inCluster bool
	httpPort  string
)

func main() {

	flag.BoolVar(&inCluster, "in_cluster", false, " bool flag if this app run inside k8s cluster (default false)")
	flag.StringVar(&httpPort, "http_port", ":8181", "define http port for resilia server")
	flag.Parse()

	var (
		kubeEngine *kube.Engine
		err        error
	)

	log.Println(inCluster)

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

	pumbaEngine := pumba.New(kubeEngine)
	suiteService := services.New(kubeEngine, pumbaEngine)

	httpAPI := httpserver.New(httpPort, 5*time.Second, suiteService)
	httpAPI.Run(httpPort)

}
