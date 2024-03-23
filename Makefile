PKG     = github.com/ArcticOJ/igloo/v0
BIN		= arctic
HASH    = $(shell git rev-parse --short HEAD)
DATE    = $(shell date +%s)
TAG     = $(shell git describe --tags --always --abbrev=0 --match="v[0-9]*.[0-9]*.[0-9]*" 2> /dev/null)
VERSION = $(shell echo "${TAG}" | sed 's/^.//')

DEFAULT_ENV = GOOS=linux

DEV_FLAGS = -ldflags "-X '${PKG}/build.Version=${VERSION}' -X '${PKG}/build.Hash=${HASH}' -X '${PKG}/build._date=${DATE}'"
REL_FLAGS = -ldflags "-X '${PKG}/build.Version=${VERSION}' -X '${PKG}/build.Hash=${HASH}' -X '${PKG}/build._date=${DATE}' -s -w"

release: main.go
	${DEFAULT_ENV} go build ${REL_FLAGS} -o ${OUT}

dev: main.go
	${DEFAULT_ENV} go build ${DEV_FLAGS} -o ${OUT}

