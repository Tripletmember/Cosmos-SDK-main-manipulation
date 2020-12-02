# for dockerized protobuf tools
PROTO_CONTAINER := cosmwasm/prototools-docker:latest
DOCKER_BUF := docker run --rm -v $(shell pwd)/buf.yaml:/workspace/buf.yaml -v $(shell go list -f "{{ .Dir }}" -m github.com/cosmos/cosmos-sdk):/workspace/cosmos_sdk_dir --workdir /workspace $(PROTO_CONTAINER)

###############################################################################
###                                Protobuf                                 ###
###############################################################################

proto-all: proto-gen proto-lint proto-check-breaking
.PHONY: proto-all

proto-gen: proto-lint
	@docker run --rm \
	  -v $(shell go list -f "{{ .Dir }}" -m github.com/cosmos/cosmos-sdk):/workspace/cosmos_sdk_dir \
	  -v $(shell go list -f "{{ .Dir }}" -m github.com/gogo/protobuf):/workspace/protobuf_dir \
	  -v $(shell pwd):/workspace \
	  --workdir /workspace \
	  --env COSMOS_SDK_DIR=/workspace/cosmos_sdk_dir \
	  --env PROTOBUF_DIR=/workspace/protobuf_dir \
	  $(PROTO_CONTAINER) ./scripts/protocgen.sh
.PHONY: proto-gen

proto-lint:
	@$(DOCKER_BUF) buf check lint --error-format=json
.PHONY: proto-lint

proto-check-breaking:
	@$(DOCKER_BUF) buf check breaking --against-input $(HTTPS_GIT)#branch=master
.PHONY: proto-check-breaking
