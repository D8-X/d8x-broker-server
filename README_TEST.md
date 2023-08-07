# DEV

### Test

`go test ./src/test/`

### Get the latest D8X Go SDK
See [how-to-use-a-private-go-module](https://www.digitalocean.com/community/tutorials/how-to-use-a-private-go-module-in-your-own-project#configuring-go-to-access-private-modules),
or [stackoverflow](https://stackoverflow.com/questions/27500861/whats-the-proper-way-to-go-get-a-private-repository),
in particular, set the "instead of credentials" in `~/.gitconfig` by adding the following line:
```
[url "https://YOURUSERNAME:YOURTOKEN@github.com/"]
        insteadOf = https://github.com/
```
and set the environment variable
```
export GOPRIVATE=github.com/D8-X/d8x-futures-go-sdk/
```
Then you can upgrade via `go get github.com/D8-X/d8x-futures-go-sdk@latest`
