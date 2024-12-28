FROM golang:1.20.3 AS build-env

ADD . /dockerdev
WORKDIR /dockerdev

RUN rm go.sum && go mod tidy

WORKDIR /dockerdev/cmd
RUN CGO_ENABLED=0 go build -o /testgenerate_subject

#Final stage
FROM alpine:3.16.0

ENV LOG_LEVEL="INFO"
ENV LISTEN_PORT=:80
ENV SECRET_KEY="secretkey"
ENV DB_HOST="192.168.12.120"
ENV DB_PORT=5432
ENV DB_NAME="generate"
ENV DB_USER="postgres"
ENV DB_PASSWORD="pgpassword"


EXPOSE 80

WORKDIR /
COPY --from=build-env /testgenerate_subject /

ENTRYPOINT ["/testgenerate_subject"]
