.PHONY: docker deploy

docker:
	docker buildx build -f deploy/Dockerfile --platform=linux/arm64,linux/amd64 --push --tag ghcr.io/xanecs/netrelay:latest $(CURDIR)

deploy:
	envsubst < deploy/kustomization.template.yaml > deploy/kustomization.yaml
	kubectl apply -k deploy