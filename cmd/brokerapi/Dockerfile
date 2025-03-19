FROM golang:1.22
ARG BUILD_VERSION
WORKDIR /broker
COPY . .

RUN go mod download && go mod verify
RUN go build -ldflags "-X 'main.VERSION=${BUILD_VERSION}'" -o /usr/local/bin/app ./cmd/brokerapi/main.go

FROM debian:bookworm-slim
COPY --from=0 /usr/local/bin/app /usr/local/bin/app
RUN apt-get update && apt-get install -y ca-certificates

CMD ["app"]
