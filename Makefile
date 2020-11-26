VERSION = v1.18.9
GIT_SHA := $(shell echo `git rev-parse --verify HEAD^{commit}`)
IMAGE_NAME = ghcr.io/kbst/kind-eks-d
TEST_IMAGE = ${IMAGE_NAME}:${GIT_SHA}

default: update-src build-image test-image

update-src:
	wget https://beta.cdn.model-rocket.aws.dev/kubernetes-1-18/releases/1/artifacts/kubernetes/${VERSION}/kubernetes-src.tar.gz
	rm -rf kubernetes-src
	mkdir kubernetes-src
	tar -xzf kubernetes-src.tar.gz -C kubernetes-src/
	rm -f kubernetes-src.tar.gz
	for name in `ls kubernetes-src/`; \
		do mv "kubernetes-src/$$name" "kubernetes-src/$${name//kubernetes/}"; \
	done

build-image:
	cd kubernetes-src; \
	KUBE_GIT_VERSION=${VERSION} kind build node-image --image ${TEST_IMAGE} --kube-root .

test-image:
	kind create cluster --name eks-d-test --image ${TEST_IMAGE}
	kubectl cluster-info
	kind delete cluster --name eks-d-test

push-image:
	docker push ${TEST_IMAGE}

pull-image:
	while true; do \
		docker pull ${TEST_IMAGE} || continue; \
		break; \
	done

promote-image:
ifndef GITHUB_REF
	$(error GITHUB_REF is not set)
endif
	RELEASE_IMAGE = ${IMAGE_NAME}:$(GITHUB_REF:refs/tags/%=%)
	docker tag ${TEST_IMAGE} ${RELEASE_IMAGE}
	docker push ${RELEASE_IMAGE}
