DOCKER   ?= docker
REGISTRY ?= hub.xdxct.com
PROJECT  ?= xdxct-docker
BASE    ?= ubuntu20.04
TAG		?= devel

.PHONY: all
all: ubuntu

push-latest:
	$(DOCKER) push $(REGISTRY)/$(PROJECT)/k8s-device-plugin:$(TAG)

clean:
	$(DOCKER) rmi $(REGISTRY)/$(PROJECT)/k8s-device-plugin:$(TAG)

ubuntu:
	$(DOCKER) build -t $(REGISTRY)/$(PROJECT)/k8s-device-plugin:$(TAG) .
