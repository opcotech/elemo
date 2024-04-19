##############################################################################
# Common image
##############################################################################

FROM golang:1.22-bullseye AS common

WORKDIR /common

RUN apt install -y --no-install-recommends git \
    && rm -rf /var/session-client/apt/lists/*

COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY . .

##############################################################################
# Build image
##############################################################################

FROM golang:1.22-bullseye AS build

ENV GOOS linux
ENV CGO_ENABLED 0

COPY --from=common /common /build
COPY --from=common $GOPATH/pkg/mod $GOPATH/pkg/mod

WORKDIR /build
RUN go build -ldflags="-s -w -extldflags=-Wl,-ld_classic" -o /usr/bin/elemo cmd/elemo/main.go

##############################################################################
# Production image
##############################################################################

FROM gcr.io/distroless/static-debian11:nonroot AS production

ENV TZ UTC

COPY --chown=nonroot:nonroot assets /elemo/assets
COPY --chown=nonroot:nonroot configs /elemo/configs
COPY --chown=nonroot:nonroot templates /elemo/templates
COPY --from=build /usr/bin/elemo /elemo/bin/elemo

WORKDIR /elemo
