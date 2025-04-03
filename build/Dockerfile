# Build stage
FROM golang:1.23-alpine3.19 AS builder

ARG VERSION
ARG BIN

RUN apk add --no-cache make

WORKDIR /go/src/github.com/EvgenyRomanov/sql-migrator

ENV GO111MODULE=on

COPY go.mod go.sum ./

RUN go mod download

COPY . ./

RUN make build-for-docker

# Final stage
FROM alpine:3.19

ARG BIN

WORKDIR /usr/bin

COPY --from=builder /go/src/github.com/EvgenyRomanov/sql-migrator/${BIN} .

# ENTRYPOINT ["gomigrator"]
# CMD ["--help"]

ENTRYPOINT ["tail"]
CMD ["-f","/dev/null"]