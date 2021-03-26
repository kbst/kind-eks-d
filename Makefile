RELEASE_BRANCH = 1-19
VERSION := v$(subst -,.,$(RELEASE_BRANCH)).6
SOURCE_URL = https://distro.eks.amazonaws.com/kubernetes-${RELEASE_BRANCH}/releases/1/artifacts/kubernetes/${VERSION}/kubernetes-src.tar.gz
GIT_SHA := $(shell echo `git rev-parse --verify HEAD^{commit}`)
IMAGE_NAME ?= ghcr.io/kbst/kind-eks-d
TEST_IMAGE = ${IMAGE_NAME}:${GIT_SHA}

default: update-src build-image test-image

update-src:
	wget ${SOURCE_URL}
	rm -rf kubernetes-src
	mkdir kubernetes-src
	tar -xzf kubernetes-src.tar.gz -C kubernetes-src/
	rm -f kubernetes-src.tar.gz

build-image:
	kind --version
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

RELEASE_IMAGE = ${IMAGE_NAME}:$(subst refs/tags/,,${GITHUB_REF})
promote-image:
ifndef GITHUB_REF
	$(error GITHUB_REF is not set)
endif
	docker tag ${TEST_IMAGE} ${RELEASE_IMAGE}
	docker push ${RELEASE_IMAGE}
