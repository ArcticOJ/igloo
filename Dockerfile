FROM --platform=${BUILDPLATFORM} golang:alpine AS builder
WORKDIR /usr/src/app

ARG TARGETOS
ARG TARGETARCH

ENV CGO_ENABLED=0

RUN apk add --no-cache make git

COPY go.mod go.sum ./
RUN go mod download

COPY .. .

RUN --mount=type=cache,target=/go/pkg/mod for variant in tier1 tier2 tier3; do GOOS=${TARGETOS} GOARCH=${TARGETARCH} make release OUT="./out/igloo.${variant}" VARIANT=${variant}; done
RUN --mount=type=cache,target=/go/pkg/mod GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o ./out/gen_runtimes ./cmd/gen_runtimes

FROM --platform=${BUILDPLATFORM} alphanecron/judge-env:tier-1 AS tier-1
WORKDIR /igloo

COPY --from=builder /usr/src/app/out/igloo.tier1 ./igloo
COPY --from=builder /usr/src/app/out/gen_runtimes ./gen_runtimes

RUN ./gen_runtimes

ENTRYPOINT ["/igloo/igloo"]

FROM --platform=${BUILDPLATFORM} alphanecron/judge-env:tier-2 AS tier-2
WORKDIR /igloo

COPY --from=builder /usr/src/app/out/igloo.tier2 ./igloo
COPY --from=builder /usr/src/app/out/gen_runtimes ./gen_runtimes

RUN ./gen_runtimes

ENTRYPOINT ["/igloo/igloo"]

#FROM --platform=${BUILDPLATFORM} alphanecron/judge-env:tier-3 AS tier-3
#WORKDIR /igloo
#
#COPY --from=builder /usr/src/app/out/igloo.tier3 ./igloo
#
#ENTRYPOINT ["/igloo/igloo"]
