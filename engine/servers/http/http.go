package http

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	suites "github.com/faruqisan/resilia/engine/suites/services"
	"github.com/gin-gonic/gin"
)

type (

	// SuitesService interface define contract with suite service
	SuitesService interface {
		NewModel(id string, name string, resources []suites.FileResource) *suites.Model
		RunSuiteFileResources(suite *suites.Model) error
		RunSuitePumbaWorkers(suite *suites.Model) error
		StopSuites(suite *suites.Model) error
	}

	// Engine struct hold http server engine required data
	Engine struct {
		port    string
		timeout time.Duration
		router  *gin.Engine

		suiteService SuitesService
	}
)

// New function return setuped http server engine
func New(port string, timeout time.Duration, suitesService SuitesService) *Engine {
	e := &Engine{
		port:         port,
		timeout:      timeout,
		router:       gin.Default(),
		suiteService: suitesService,
	}

	e.initRoutes()
	return e
}

// Run function run new http server
func (e *Engine) Run(port string) {

	apiServer := &http.Server{
		Addr:    e.port,
		Handler: e.router,
	}

	go func() {
		// service connections
		if err := apiServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown Server ...")
	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()
	if err := apiServer.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown: ", err)
	}
	log.Println("Server exiting")

}
