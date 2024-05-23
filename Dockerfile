FROM golang:1.22 AS builder

ENV GO111MODULE=on \
  CGO_ENABLED=0 \
  GOOS=linux \
  GOARCH=amd64

RUN apt-get -qq update

WORKDIR /src
COPY . .

RUN go build \
  -ldflags "-s -w -extldflags '-static'" \
  -o /bin/app \
  . \
  && strip /bin/app

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /bin/app /forkdiff

ENTRYPOINT ["/forkdiff"]
