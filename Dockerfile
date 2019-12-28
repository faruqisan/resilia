FROM golang:alpine AS builder

LABEL maintainer="Ikhsan Faruqi <faruqisan@gmail.com>"

WORKDIR /app
ADD . /app
RUN cd /app & go mod download
RUN cd /app & go build -o resilience_k8s cmd/main.go

FROM alpine
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
WORKDIR /app
COPY --from=builder /app/resilience_k8s /app
COPY --from=builder /app/files /app/files

ENTRYPOINT ./resilience_k8s