FROM golang:1.26.2-trixie AS build

WORKDIR /src

COPY go.* ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 go build ./cmd -v -o c

FROM debian:trixie-slim

RUN set -x && \
    apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y ca-certificates && \
    rm -rf /var/lib/apt/lists/*

COPY --from=build /src/kith-pms /usr/bin/kith-pms

CMD ["/usr/bin/kith-pms"]
