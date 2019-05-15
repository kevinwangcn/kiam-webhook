FROM golang:1.11 AS builder
LABEL authors="Mattia Mascia <mmascia@redhat.com>"

ADD https://github.com/golang/dep/releases/download/v0.5.3/dep-linux-amd64 /usr/bin/dep
RUN chmod +x /usr/bin/dep

WORKDIR $GOPATH/src/github.com/openlab-red/kiam-webhook

COPY Gopkg.toml Gopkg.lock ./

RUN dep ensure --vendor-only

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix nocgo -o /app .

FROM registry.access.redhat.com/ubi7/ubi:latest

ENV HOME=/home/kiam-webhook
RUN mkdir -p $HOME

COPY --from=builder /app $HOME

RUN chown -R 1001:0 $HOME && \
    chmod -R g+rw $HOME

WORKDIR $HOME

USER 1001

ENTRYPOINT ["./app"]