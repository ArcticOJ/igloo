# TODO: Add Windoze support

FROM golang:alpine AS builder
WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o ./out/polar -ldflags "-s -w" main.go

FROM golang:alpine AS judge-env

RUN apk add gcc python3 fpc pypy3 clang --no-cache \
  --repository https://dl-cdn.alpinelinux.org/alpine/edge/testing \
  --repository https://dl-cdn.alpinelinux.org/alpine/edge/main

FROM judge-env AS runner
WORKDIR /polar

EXPOSE 172/tcp

COPY --from=builder /usr/src/app/out/polar ./

ENTRYPOINT ["/polar/polar"]

