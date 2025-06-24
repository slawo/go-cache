FROM mcr.microsoft.com/devcontainers/go:1.24
RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
  && sudo apt-get -y install git protobuf-compiler
RUN go install github.com/vektra/mockery/v2@v2.53.4
RUN chown -R vscode /go
