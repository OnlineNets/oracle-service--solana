FROM golang:1.17-alpine as buildbase

RUN apk add git build-base

WORKDIR /go/src/gitlab.com/tokend/enrex/vesting-cleaner

COPY . .

ENV GO111MODULE="on"
ENV CGO_ENABLED=0
ENV GOOS="linux"

RUN go mod vendor
RUN go build -a -mod=vendor -o /usr/local/bin/vesting-cleaner gitlab.com/tokend/enrex/vesting-cleaner


###

FROM alpine:3.9

COPY --from=buildbase /usr/local/bin/vesting-cleaner /usr/local/bin/vesting-cleaner
RUN apk add --no-cache ca-certificates

ENTRYPOINT ["vesting-cleaner"]
