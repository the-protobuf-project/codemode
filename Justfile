# codemode Justfile
# Run `just --list` to see all available recipes.

bin := "protoc-gen-codemode"
version := "v0.1.0"

# Default: list all recipes
default:
    @just --list

# Build the protoc-gen-codemode plugin with its version baked in
build:
    go build -trimpath -ldflags "-X main.version={{version}}" -o ./bin/{{bin}} ./plugin/cmd/protoc-gen-codemode/

# Cross-compile for a specific OS/ARCH
build-cross os arch:
    GOOS={{os}} GOARCH={{arch}} go build -trimpath -ldflags "-s -w -X main.version={{version}}" \
        -o ./bin/{{bin}}-{{os}}-{{arch}}{{if os == "windows" { ".exe" } else { "" } }} \
        ./plugin/cmd/protoc-gen-codemode/

# Install the plugin to $GOPATH/bin
install:
    go install -ldflags "-X main.version={{version}}" ./plugin/cmd/protoc-gen-codemode/

# Regenerate the Go types for the codemode.v1 annotations
proto:
    cd protobuf && buf generate

# Regenerate the library example
generate: build
    rm -rf examples/library/gen
    buf generate --template examples/library/buf.gen.yaml

# Lint protos, Go, TypeScript (Biome) and the 200-line limit
lint:
    buf lint
    go vet ./...
    bun run lint
    ./scripts/check-lines.sh

# Run Go tests, then type-check and test the example on Bun
test:
    go test ./...
    bun run typecheck
    bun run test

# Run the example's tests on Node and Deno too
test-runtimes: test
    bun run test:node
    bun run test:deno

# Fail if regenerating changes the committed example
check-generated: generate
    git diff --exit-code -- examples/library/gen
