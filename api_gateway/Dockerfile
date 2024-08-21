FROM golang:latest AS build

WORKDIR /app

COPY go.mod ./

RUN go mod tidy

COPY . .

RUN CG0_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o gateway ./cmd/gateway

FROM alpine:latest

WORKDIR /app

COPY --from=build /app/gateway .

ENTRYPOINT ["./gateway"]