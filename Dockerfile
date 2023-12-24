FROM --platform=${BUILDPLATFORM} golang:alpine AS builder
WORKDIR /usr/src/app

ARG TARGETOS
ARG TARGETARCH

ENV CGO_ENABLED=0

COPY go.mod go.sum ./
RUN go mod download

COPY .. .

RUN --mount=type=cache,target=/go/pkg/mod --mount=type=bind,target=. for variant in "tier1 tier2 tier3"; do GOOS=${TARGETOS} GOARCH=${TARGETARCH} make release OUT="./out/igloo.${variant}" VARIANT=${variant}; done

FROM --platform=${BUILDPLATFORM} alphanecron/judge-env:tier-1 AS igloo-tier-1
WORKDIR /igloo

COPY --from=builder /usr/src/app/out/igloo.tier1 ./igloo

ENTRYPOINT ["/igloo/igloo"]

FROM --platform=${BUILDPLATFORM} alphanecron/judge-env:tier-2 AS igloo-tier-2
WORKDIR /igloo

COPY --from=builder /usr/src/app/out/igloo.tier2 ./igloo

ENTRYPOINT ["/igloo/igloo"]

#FROM --platform=${BUILDPLATFORM} alphanecron/judge-env:tier-3 AS igloo-tier-3
#WORKDIR /igloo
#
#COPY --from=builder /usr/src/app/out/igloo.tier3 ./igloo
#
#ENTRYPOINT ["/igloo/igloo"]
