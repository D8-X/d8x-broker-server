FROM golang:1.21
WORKDIR /broker
COPY . . 

#RUN apt update
RUN go mod download && go mod verify
RUN go build -o /usr/local/bin/app ./cmd/executorws/main.go

FROM debian:bookworm
COPY --from=0 /usr/local/bin/app /usr/bin/app
CMD ["app"]
