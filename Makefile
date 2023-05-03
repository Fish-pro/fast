GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
VERSION = "latest"
REGISTRY = "fishpro3/fast"

TARGETS := fast-controller-manager \
			fast-agent

CTL_TARGETS := fast

CMD_TARGET=$(TARGETS) $(CTL_TARGETS)

.PHONY: $(CMD_TARGET)
$(CMD_TARGET):
	BUILD_PLATFORMS=$(GOOS)/$(GOARCH) hack/build.sh $@

IMAGE_TARGET=$(addprefix image-, $(TARGETS))
.PHONY: $(IMAGE_TARGET)
$(IMAGE_TARGET):
	set -e;\
	target=$$(echo $(subst image-,,$@));\
	make $$target GOOS=linux;\
	VERSION=$(VERSION) REGISTRY=$(REGISTRY) BUILD_PLATFORMS=linux/$(GOARCH) hack/docker.sh $$target

images: $(IMAGE_TARGET)

protoc:
	 cd api/proto/v1 && protoc --go_out=. --go_opt=paths=source_relative \
        --go-grpc_out=. --go-grpc_opt=paths=source_relative \
        *.proto

codegen:
	hack/update-codegen.sh

crdgen:
	hack/update-crdgen.sh
