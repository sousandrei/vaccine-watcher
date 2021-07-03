IMAGE=vaccine
REGISTRY_URL=gcr.io/sousandrei
VERSION=$(shell git rev-parse --short=7 HEAD)


build:
	docker build . -t ${IMAGE}

push:
	docker tag ${IMAGE} ${REGISTRY_URL}/${IMAGE}:${VERSION}
	docker push ${REGISTRY_URL}/${IMAGE}:${VERSION}

deploy-ci:
	helm upgrade vaccine ./chart --namespace vaccine --set image=${VERSION}

deploy: build push deploy-ci