# d8x-broker-server

server to be used as remote broker by trader back-end

## Run
Build `go build cmd/main.go`

You can then build and run the Docker image:
```
$ docker build -t broker-server . -f cmd/Dockerfile
$ dockerun -p 8000:8000 broker-server
```
note: as long as the repo is private, use:
```
$ docker build -t broker-server . -f cmd/Dockerfile --build-arg GITHUB_USER=<youruser> --build-arg GITHUB_TOKEN=<yourtoken>
$ dockerun -p 8000:8000 broker-server
```

# Endpoints (WIP!)

GET: /broker-address

`{"brokerAddr":"0x5A09217F6D36E73eE5495b430e889f8c57876Ef3"}`

GET: /broker-fee
`{"BrokerFeeTbps":60}`

POST: /sign-order

```
{
    "order": {
        "iDeadline": 1688347462,
        "traderAddr": "0x9d5aaB428e98678d0E645ea4AeBd25f744341a05",
        "iPerpetualId": 10001},
    "chainId": 80001
}
```

Response:

```
{
    "orderFields": {
        "iPerpetualId": 10001,
        "brokerFeeTbps": 60,
        "brokerAddr": "0xabcedef0123456789abcde..",
        "traderAddr": "0x9d5aaB428e98678d0E645ea4AeBd25f744341a05",
        "iDeadline": 1688347462
    },
    "chainId": 80001,
    "brokerSignature": "0x73ecb2d9ccd577b441333bb9d5fcd9a625cd2fdef5203d0b9808befab0e7e02053f8e0deac0602f1cc294f4706281f83a48745cee92a7bf61cef0516ec7514f21b"
}
```
POST: sign-payment
```
{
    "payment": {
        "payer": "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266", 
        "executor": "0x70997970C51812dc3A010C7d01b50e0d17dc79C8", 
        "token": "0xe7f1725E7734CE288F8367e1Bb143E90bb3F0512", 
        "timestamp": 1688518335, 
        "id": 42,
        "totalAmount": 1000
    },
    "chainId": 80001,
    "signature": "0x3ba983fd03c309252904d8fb8fb49943a89698fb28545df2cc3cb581a19272ac0875fa23c4b617b9c7dde41553f5a9ef38896358bd3c36f983357fde4336c4f61b"
}
```
Response:
```
{"BrokerSignature": "0x..."},
```
