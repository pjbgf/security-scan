# Define VERSION, which is used for image tags or to bake it into the
# compiled binary to enable the printing of the application version, 
# via the --version flag.
CHANGES = $(shell git status --porcelain --untracked-files=no)
ifneq ($(CHANGES),)
    DIRTY = -dirty
endif

# Prioritise DRONE_TAG for backwards compatibility. However, the git tag
# command should be able to gather the current tag, except when the git
# clone operation was done with "--no-tags".
ifneq ($(DRONE_TAG),)
	GIT_TAG = $(DRONE_TAG)
else
	GIT_TAG = $(shell git tag -l --contains HEAD | head -n 1)
endif

COMMIT = $(shell git rev-parse --short HEAD)
VERSION = $(COMMIT)$(DIRTY)

# Override VERSION with the Git tag if the current HEAD has a tag pointing to
# it AND the worktree isn't dirty.
ifneq ($(GIT_TAG),)
	ifeq ($(DIRTY),)
		VERSION = $(GIT_TAG)
	endif
endif

# Statically link the binary, unless when building in Darwin.
ifneq ($(shell uname -s), Darwin)
	LINKFLAGS = -extldflags -static -w -s
endif

RUNNER := docker
IMAGE_BUILDER := $(RUNNER) buildx

ifeq ($(TAG),)
	TAG = $(VERSION)
	ifneq ($(DIRTY),)
		TAG = dev
	endif
endif

GO := go

# Define the target platforms that can be used across the ecosystem.
# Note that what would actually be used for a given project will be
# defined in TARGET_PLATFORMS, and must be a subset of the below:
DEFAULT_PLATFORMS := linux/amd64,linux/arm64,linux/x390s,linux/riscv64

.PHONY: help
help: ## display Makefile's help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

buildx-machine: ## create rancher dockerbuildx machine targeting platform defined by DEFAULT_PLATFORMS.
	@docker buildx ls | grep rancher || \
		docker buildx create --name=rancher --platform=$(DEFAULT_PLATFORMS) --use

## Optional
# docker buildx imagetools inspect ghcr.io/pjbgf/security-scan:v0.0.1-rc6 --format "{{ json .SBOM }}"
# docker buildx imagetools inspect ghcr.io/pjbgf/security-scan:v0.0.1-rc6 --format "{{ json .Provenance }}"
BUILDX_ARGS ?= --attest type=sbom --attest type=provenance,mode=max
FULCIO_URL ?= https://fulcio.sigstore.dev
REKOR_URL ?= https://rekor.sigstore.dev

COSIGN = $(TOOLS_BIN)/cosign
$(COSIGN): ## Download cosign locally if not yet downloaded.
	$(call go-install-tool,$(COSIGN),github.com/sigstore/cosign/v2/cmd/cosign@latest)

image-push-and-sign: IID_FILE=$(shell mktemp) ## push then sign image using cosign.
image-push-and-sign:
	$(MAKE) hack-push-and-sign IID_FILE=$(IID_FILE)

hack-push-and-sign: $(COSIGN)
ifeq ($(IID_FILE),)
	@echo "invalid target, use image-push-and-sign instead"; exit 1
endif
	$(MAKE) image-push IID_FILE_FLAG="--iidfile $(IID_FILE)"
	$(COSIGN) sign --yes "$(REPO)/security-scan@$$(head -n 1 $(IID_FILE))" \
		--oidc-provider=github-actions \
		--fulcio-url=$(FULCIO_URL) --rekor-url=$(REKOR_URL)
	rm -f $(IID_FILE)
