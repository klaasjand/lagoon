FROM golang:1.17.0-buster
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
RUN go get -v golang.org/x/tools/cmd/goimports
# Simple utility for extracting a JSON representation of the declarations in a Go source file.
RUN go get -v github.com/ramya-rao-a/go-outline
# gopls provides IDE features, such as as intelligent autocompletion, signature help, refactoring, and workspace symbol search.
RUN go get -v golang.org/x/tools/gopls

# Below tools are needed for the basic features of the Go extension.
RUN go get -v github.com/uudashr/gopkgs/v2/cmd/gopkgs
RUN go get -v github.com/go-delve/delve/cmd/dlv
RUN go get -v honnef.co/go/tools/cmd/staticcheck
