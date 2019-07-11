FROM golang:1.12.7-stretch AS gobuild

WORKDIR /go/src/github.com/distributed-containers-inc/knoci
ENV GO111MODULE on
ENV CGO_ENABLED 0
ENV GOOS linux

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY pkg ./pkg
RUN go build -mod=readonly -o ./app cmd/operator/main.go

FROM scratch
COPY --from=gobuild /go/src/github.com/distributed-containers-inc/knoci/app /
ENTRYPOINT ["/app"]
