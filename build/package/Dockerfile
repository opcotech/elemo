##############################################################################
# Common image
##############################################################################

FROM golang:1.26-trixie AS common

WORKDIR /common

RUN apt-get update \
    && apt-get install -y --no-install-recommends git \
    && rm -rf /var/lib/apt/lists/*

COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY . .

##############################################################################
# Build image
##############################################################################

FROM common AS build

ENV GOOS=linux
ENV CGO_ENABLED=0

WORKDIR /common
RUN go build -ldflags="-s -w -extldflags=-Wl,-ld_classic" -o /usr/bin/elemo cmd/elemo/main.go
RUN mkdir -p /skeleton/plugins

##############################################################################
# Production image
##############################################################################

FROM gcr.io/distroless/static-debian13:nonroot AS production

ENV TZ=UTC

WORKDIR /opt/opcotech/elemo

COPY --chown=nonroot:nonroot LICENSE LICENSE-COMMERCIAL /opt/opcotech/elemo/
COPY --chown=nonroot:nonroot assets/keys /opt/opcotech/elemo/assets/keys
COPY --chown=nonroot:nonroot templates/email /opt/opcotech/elemo/templates/email
COPY --chown=nonroot:nonroot --from=build /usr/bin/elemo /opt/opcotech/elemo/bin/elemo

# Plugin installs are written at runtime; a named volume mounted here inherits
# this ownership, so the nonroot process can replace existing versions.
COPY --chown=nonroot:nonroot --from=build /skeleton/plugins /opt/opcotech/elemo/plugins

CMD ["bin/elemo", "start", "server"]
