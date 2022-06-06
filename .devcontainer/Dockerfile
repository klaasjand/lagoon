FROM golang:1.18.3-buster
LABEL maintainer="Klaas Jan Dijksterhuis"

RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install --no-install-recommends \
        curl \
        tree \
        rsync \
        debmirror \
        yum-utils \
        createrepo

# goimports updates your Go import lines, adding missing ones and removing unreferenced ones.
RUN go install golang.org/x/tools/cmd/goimports@latest
# Simple utility for extracting a JSON representation of the declarations in a Go source file.
RUN go install github.com/ramya-rao-a/go-outline@latest
# gopls provides IDE features, such as as intelligent autocompletion, signature help, refactoring, and workspace symbol search.
RUN go install golang.org/x/tools/gopls@latest

# Below tools are needed for the basic features of the Go extension.
RUN go install github.com/uudashr/gopkgs/v2/cmd/gopkgs@latest
RUN go install github.com/go-delve/delve/cmd/dlv@latest
RUN go install honnef.co/go/tools/cmd/staticcheck@latest
