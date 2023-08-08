# d8x-broker-server

server to be used as remote broker by trader back-end

## Run
Copy configuration and edit allowedExecutors (addresses which are permissioned to execute payments):
`cp config/example.chainConfig.json config/live.chainConfig.json` and
`nano config/live.chainConfig.json` to edit.
Build with `go build cmd/main.go`

You can then build and run the Docker image:
```
$ docker build -t broker-server . -f cmd/Dockerfile
$ docker run -p 8000:8000 broker-server
```
note: as long as the repo is private, use:
```
$ docker build -t broker-server . -f cmd/Dockerfile --build-arg GITHUB_USER=<youruser> --build-arg GITHUB_TOKEN=<yourtoken>
$ docker run -p 8000:8000 broker-server
```

# Endpoints

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
        "payer": "0x4Fdc785fe2C6812960C93CA2F9D12b5Bd21ea2a1", 
        "executor": "0xDa47a0CAc77D50114F2725D06a2Ce887cF9f4D98", 
        "token": "0x2d10075E54356E16Ebd5C6BB5194290709B69C1e", 
        "timestamp": 1691249493, 
        "id": 1,
        "totalAmount": 1000000000000000000,
        "chainId": 80001,
        "multiPayCtrct": "0x30b55550e02B663E15A95B50850ebD20363c2AD5"
    },
    "signature": "0x368e159104505a22a8bef736d0bbd190ffdeaa9030d76841e831082d4b0469ce22a034ed9672dd88324e22f479b08aa5c6729d3319c1d3db1b535068d86866571c"
}
```
Response:
```
{"BrokerSignature": "0x..."},
```
Errors:
1. `{"error":"wrong signature"}`
The executor has to sign the payment data cryptographically. If the executor-address cannot be recovered from the signed data,
this is the error.

2. `{"error":"executor not allowed"}`

Executors are permissioned in `live.chainConfig.json`

