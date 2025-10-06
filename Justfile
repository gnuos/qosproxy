alias b := build
alias c := clean

program := "qosproxy"

export HTTPS_PROXY := "$https_proxy"

default:
    @just build

build:
    @go fmt .
    @go mod tidy
    CGO_ENABLE=0 GO111MODULE=on go build -buildvcs=false -gcflags="all=-N -l" -ldflags='-extldflags "-static" -w -s -buildid=' -trimpath -o bin/{{program}} .

image:
    buildah build --build-arg PROGRAM_NAME={{program}} --tag {{program}}:latest .
    buildah prune -f

clean:
    rm -f ./bin/*

