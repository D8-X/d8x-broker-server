FROM golang:1.21
WORKDIR /broker
COPY . . 
COPY .env .env

#RUN apt update
RUN go mod download && go mod verify
RUN go build -o /usr/local/bin/app ./cmd/executorws/main.go

CMD ["app"]
