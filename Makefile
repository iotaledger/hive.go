PROTOC_GEN_GO := $(GOPATH)/bin/protoc-gen-go
PROTO_FILES := $(shell find . -path ./vendor -prune -o -type f -name '*.proto' -print)
PROTO_GO_FILES := $(patsubst %.proto, %.pb.go, $(PROTO_FILES))

# generate all proto files
.PHONY: compile
compile: $(PROTO_GO_FILES)

# if $GOPATH/bin/protoc-gen-go does not exist, install it
$(PROTOC_GEN_GO):
	go install google.golang.org/protobuf/cmd/protoc-gen-go

# implicit compile rule for proto files
%.pb.go: %.proto | $(PROTOC_GEN_GO)
	protoc $< --plugin=$(PROTOC_GEN_GO) --go_out=paths=source_relative:.
