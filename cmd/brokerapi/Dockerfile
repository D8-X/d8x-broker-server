FROM golang:1.21
WORKDIR /broker
COPY . . 

#RUN apt update
RUN go mod download && go mod verify
RUN go build -o /usr/local/bin/app ./cmd/brokerapi/main.go

CMD ["app"]
