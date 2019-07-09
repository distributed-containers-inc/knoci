FROM golang:1.12.7-stretch AS gobuild

WORKDIR /go/src/github.com/distributed-containers-inc/knoci
ENV GO111MODULE on

COPY go.mod go.sum ./
RUN go get -d -v ./...
COPY main.go ./
RUN go build -o /app main.go

FROM scratch
COPY --from=gobuild /app /app
ENTRYPOINT ["/app"]
