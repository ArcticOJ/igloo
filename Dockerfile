# TODO: implement tiered docker images
# tier 1: gcc python3 python2 clang fpc pypy3 go
# tier 2: more runtimes such as kotlin, java, csharp, etc..
# tier 3: rarely used languages like brainfuck, whitespace, moo, ...

FROM golang:alpine AS builder
WORKDIR /usr/src/app

ARG CGO_ENABLED=0

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o ./out/igloo -ldflags "-s -w" main.go

FROM golang:bullseye AS judge-env

ARG DEBIAN_FRONTEND=noninteractive

RUN apt update && apt install -y gcc python3 python2 clang fpc pypy3 && rm -rf /var/lib/apt/lists/*

FROM judge-env AS runner
WORKDIR /igloo

EXPOSE 172/tcp

COPY --from=builder /usr/src/app/out/igloo ./

ENTRYPOINT ["/igloo/igloo"]

