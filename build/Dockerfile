# Build golang app
FROM golang:1.20.1-buster AS builder

WORKDIR /go/src/github.com/klaasjand/lagoon

COPY ./ ./

RUN go mod tidy \
    && go mod verify \
    && go mod download

RUN CGO_ENABLED=0 GOARCH=$TARGETARCH GOOS=$TARGETOS go build -a -o lagoon ./cmd

# Create final Docker images
FROM debian:buster-slim
LABEL maintainer="Klaas Jan Dijksterhuis"

RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get install -y --no-install-recommends \
        rsync \
        debmirror \
        yum-utils \
        createrepo \
    && rm -rf /var/lib/apt/lists/*

EXPOSE 9000/tcp

RUN mkdir -p /var/lib/lagoon

WORKDIR /etc/lagoon

COPY --from=builder /go/src/github.com/klaasjand/lagoon/lagoon ./
CMD ["./lagoon"]
