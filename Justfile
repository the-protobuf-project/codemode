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

# Lint protos, Go and the 200-line limit
lint:
    buf lint
    go vet ./...
    ./scripts/check-lines.sh

# Run Go tests, then the TypeScript tests on Node
test:
    go test ./...
    npm run typecheck
    npm test

# Run the TypeScript tests on Bun and Deno too
test-runtimes: test
    npm run test:bun --workspaces --if-present
    npm run test:deno --workspaces --if-present

# Fail if regenerating changes the committed example
check-generated: generate
    git diff --exit-code -- examples/library/gen
