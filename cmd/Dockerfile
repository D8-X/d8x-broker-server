# ====== Build ======
FROM golang:1.20 AS stage-build
# see https://hub.docker.com/_/golang
# and https://levelup.gitconnected.com/golang-dockerfile-for-project-with-private-dependencies-using-https-without-ssh-9a9f0bbeddd8

# git private repo stuff
# ARG GITHUB_TOKEN
# ARG GITHUB_USER
# ENV GITHUB_TOKEN=${GITHUB_TOKEN}
# ENV GITHUB_USER=${GITHUB_USER}
# RUN go env -w GOPRIVATE=github.com/D8-X
# RUN git config --global url."https://${GITHUB_USER}:${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/"

# build
WORKDIR /app
# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./

RUN go mod download && go mod verify
COPY . .
RUN go build -o executable ./cmd/main.go

# ====== Release ======
FROM golang:1.20 AS stage-release
COPY --from=stage-build /app/executable /

# Expose the provided API_PORT
EXPOSE ${API_PORT}
ENTRYPOINT ["/executable"]
