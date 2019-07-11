FROM golang:1.12.7-stretch AS gobuild

WORKDIR /go/src/github.com/distributed-containers-inc/knoci
ENV GO111MODULE on
ENV CGO_ENABLED 0
ENV GOARCH amd64
ENV GOOS linux

COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./
COPY pkg ./pkg
RUN go build && mv ./knoci /knoci

FROM scratch
COPY --from=gobuild /knoci /
ENTRYPOINT ["/knoci"]
