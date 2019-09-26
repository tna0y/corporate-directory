FROM golang:1.12-stretch

FROM golang

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN go build cmd/server.go

RUN go test ./...

ENTRYPOINT ["/app/server"]

