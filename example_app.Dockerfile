FROM golang:alpine AS builder

LABEL maintainer="Ikhsan Faruqi <faruqisan@gmail.com>"

WORKDIR /app
ADD . /app
RUN cd /app & go mod download
RUN cd /app & go build -o resilia_example example/main.go

FROM alpine
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
WORKDIR /app
COPY --from=builder /app/resilia_example /app
COPY --from=builder /app/example/files /app/files

ENTRYPOINT ./resilia_example --in_cluster true