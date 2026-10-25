BINARY_NAME=mtb

build:
	go build -o $(BINARY_NAME) ./cmd/mtb

install: build
	cp $(BINARY_NAME) /usr/local/bin/

generate:
	go generate ./...

# spec.json を公式定義（gen-spec の sourceCommit）から作り直し、カタログも描き直す。
# 公式定義をネットワークから取る（2回目からはキャッシュ）。
spec:
	go run ./internal/gen-spec -out internal/spec/spec.json
	go generate ./...

test:
	go test ./...

# 提出前の一式。生成物のズレも含めて CI と同じことを見る。
check: generate
	gofmt -l .
	go vet ./...
	go build ./...
	go test ./...
	git diff --exit-code

clean:
	rm -f $(BINARY_NAME)
	rm -rf dist/

.PHONY: build install generate spec test check clean
