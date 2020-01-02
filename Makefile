run:
	go run cmd/main.go

deploy:
	kubectl apply -f files/k8s/deployment.yaml

run_example:
	cd example && go run main.go

deploy_example:
	docker build -f example_app.Dockerfile . -t faruqisan/resilia:0.0.1
	kubectl run resilience-k8s --image=faruqisan/resilia:0.0.1 --image-pull-policy=Never

cleanup:
	kubectl delete deployments resilience-k8s