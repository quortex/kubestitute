# Shell used by default (usefull for CI)
SHELL := /bin/bash

# unescape tpl capture text inside <tpl tpl> tag
# used to escape invalid yaml syntax for helm through kustomize
define unescape-tpl
sed 's/<tpl \(.*\) tpl>/\1/'
endef

all: chart docs

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


# Build helm chart templates from config
.PHONY: chart
chart:
	@$(KUSTOMIZE) build ../config/crd | $(call unescape-tpl) > ./kubestitute/crds/crds.yaml
	@$(KUSTOMIZE) build ./config/rbac | $(call unescape-tpl) > ./kubestitute/templates/rbac.yaml

# Build documentation
.PHONY: docs
docs:
	@docker run --rm --volume "$$(pwd)/kubestitute:/helm-docs" jnorwood/helm-docs:latest -s file
