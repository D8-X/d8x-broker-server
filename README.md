# d8x-broker-server

server to be used as remote broker by trader back-end

`go build cmd/main.go`

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
