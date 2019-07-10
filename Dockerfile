FROM golang:1.12.7-stretch AS gobuild

WORKDIR /go/src/github.com/distributed-containers-inc/knoci
ENV GO111MODULE on
ENV CGO_ENABLED 0

COPY go.mod go.sum ./
RUN go get -d -v cmd/operator/main.go
COPY cmd ./cmd
COPY pkg ./pkg
RUN go build -o ./app cmd/operator/main.go && chmod 755 ./app

FROM scratch
COPY --from=gobuild /go/src/github.com/distributed-containers-inc/knoci/app /
ENTRYPOINT ["/app"]
