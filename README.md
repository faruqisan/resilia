# resilia

Resilia is an helper service to automate setup resilience testing on kubernetes cluster.

## Setup

### Requirements

- minikube (run on local)
- kubernetes cluster
- kubernetes service account with privilege to manage cluster resource

### Run example in local (in cluster)

```bash
$ minikube start
$ eval $(minikube docker-env)
$ make deploy_example
$ make cleanup # for stop
$ kubectl get pods
```

### Run example in local (outside cluster)

```bash
$ minikube start
$ make run_example
```

you should see this on your terminal when running example app
```bash
NAME                                READY   STATUS    RESTARTS   AGE
redis-deployment-b467466b5-mgjgd    1/1     Running   0          17s
redis-deployment-b467466b5-rxq96    1/1     Running   0          17s
resilience-k8s-5486599977-mjp68     1/1     Running   0          18s
resilience-pumba-2742331732-frqsx   1/1     Running   0          17s
```

### Run Resilia in local

```bash
$ make run
```

### Run Resilia in Production

```bash
$ kubectl apply -f files/k8s/deployment.yaml
```

## TODO

- Create web service for manage the app
- Unit test. currently the coverage is 0