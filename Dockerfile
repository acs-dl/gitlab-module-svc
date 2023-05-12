FROM golang:1.18-alpine as buildbase

RUN apk add git build-base

WORKDIR /go/src/github.com/acs-dl/gitlab-module-svc
COPY vendor .
COPY . .

RUN GOOS=linux go build  -o /usr/local/bin/gitlab-module /go/src/github.com/acs-dl/gitlab-module-svc


FROM alpine:3.9

COPY --from=buildbase /usr/local/bin/gitlab-module /usr/local/bin/gitlab-module
RUN apk add --no-cache ca-certificates

ENTRYPOINT ["gitlab-module"]
