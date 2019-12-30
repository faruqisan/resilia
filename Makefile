deploy:
	docker build -t faruqisan/resilia:0.0.1 .
	kubectl run resilience-k8s --image=faruqisan/resilia:0.0.1 --image-pull-policy=Never

cleanup:
	kubectl delete deployments resilience-k8s