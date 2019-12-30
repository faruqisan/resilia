# resilia

Resilia is an helper service to automate setup resilience testing on kubernetes cluster.

## Setup

### Requirements

- minikube (run on local)
- kubernetes cluster
- kubernetes service account with privilege to manage cluster resource

### Run in local

```bash
$ minikube start
$ eval $(minikube docker-env)
$ make deploy
$ make cleanup # for stop
```

### Run in Production

```bash
$ kubectl apply -f k8s/deployment.yaml
```

## TODO

- move current cmd/main.go to example app folder
- Create web service for manage the app
- Unit test. currently the coverage is 0