FROM golang:1.20.2-alpine AS builder

WORKDIR /workspace

RUN apk add --update --no-cache git && rm -rf /var/cache/apk/*
COPY go.mod go.sum /workspace/
RUN go mod download
COPY main.go /workspace/
RUN go build -o api .

FROM alpine
RUN apk add --update --no-cache ca-certificates tzdata && rm -rf /var/cache/apk/*
COPY --from=builder /workspace/api /usr/local/bin/api
CMD [ "/usr/local/bin/api" ]
