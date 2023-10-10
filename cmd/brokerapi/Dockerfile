FROM golang:1.21
WORKDIR /broker
COPY . . 
COPY ./config/keyfile.txt /mydata/key_config/keyfile.txt
RUN rm ./config/keyfile.txt

RUN go mod download && go mod verify
RUN go build -o /usr/local/bin/app ./cmd/brokerapi/main.go

CMD ["app"]
