
# Image URL to use all building/pushing image targets
IMG ?= quortexio/kubestitute:latest
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# Shell used by default (usefull for CI)
SHELL := /bin/bash

# Test binaries repo
ENVTEST_ASSETS_DIR=$(shell pwd)/testbin

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: manager manifests

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.5 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

# find or download kustomize
# download kustomize if necessary
kustomize:
ifeq (, $(shell which kustomize))
	@{ \
	set -e ;\
	KUSTOMIZE_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$KUSTOMIZE_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/kustomize/kustomize/v3@v3.5.4 ;\
	rm -rf $$KUSTOMIZE_GEN_TMP_DIR ;\
	}
KUSTOMIZE=$(GOBIN)/kustomize
else
KUSTOMIZE=$(shell which kustomize)
endif

# find or download crd-ref-docs
# download crd-ref-docs if necessary
crd-ref-docs:
ifeq (, $(shell which crd-ref-docs))
	@{ \
	set -e ;\
	CRDREFDOCS_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CRDREFDOCS_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get github.com/elastic/crd-ref-docs@v0.0.5 ;\
	rm -rf $$CRDREFDOCS_GEN_TMP_DIR ;\
	}
CRD_REF_DOCS=$(GOBIN)/crd-ref-docs
else
CRD_REF_DOCS=$(shell which crd-ref-docs)
endif

# Run tests
test: generate fmt vet manifests
	mkdir -p "${ENVTEST_ASSETS_DIR}/" > /dev/null
		test -f ${ENVTEST_ASSETS_DIR}/setup-envtest.sh || \
			curl -sSLo ${ENVTEST_ASSETS_DIR}/setup-envtest.sh https://raw.githubusercontent.com/kubernetes-sigs/controller-runtime/master/hack/setup-envtest.sh
		source ${ENVTEST_ASSETS_DIR}/setup-envtest.sh; \
			fetch_envtest_tools $(ENVTEST_ASSETS_DIR); \
			setup_envtest_env $(ENVTEST_ASSETS_DIR); \
	go test ./... -coverprofile cover.out

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run ./main.go

# Install CRDs into a cluster
install: manifests kustomize
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

# Uninstall CRDs from a cluster
uninstall: manifests kustomize
	$(KUSTOMIZE) build config/crd | kubectl delete -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests kustomize
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | kubectl apply -f -

# Destroy controller in the configured Kubernetes cluster in ~/.kube/config
destroy: manifests kustomize
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | kubectl delete -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

# Build the docker image
docker-build: test
	docker build . -t ${IMG}

# Push the docker image
docker-push:
	docker push ${IMG}

# Generate clients
# Generates client from swagger documentation.
.PHONY: clients
clients:
	@t=$$(mktemp -d) && \
		cd $${t} && \
		git clone -b develop git@github.com:quortex/aws-ec2-adapter.git && \
		cd - && \
		cp $${t}/aws-ec2-adapter/docs/swagger.yaml ./clients/ec2adapter && \
		go generate ./...

# Build documentation
.PHONY: docs
docs: crd-ref-docs
	$(CRD_REF_DOCS) --source-path=api \
		--renderer=asciidoctor \
		--config=hack/doc-generation/config.yaml \
		--templates-dir=hack/doc-generation/templates/asciidoctor \
		--output-path=docs/api-docs.asciidoc
