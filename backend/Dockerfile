FROM golang:1.24-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git
COPY . .
RUN go mod download
RUN go build -mod=mod -o gomotz ./cmd/main.go
RUN git clone https://github.com/mascarenhasmelson/tcp-tunnel-go.git /tcp-tunnel-go \
    && cd /tcp-tunnel-go \
    && go mod tidy \
    && go build -o /tcp
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/gomotz .
COPY --from=builder /tcp ./tcp
RUN chmod +x gomotz tcp
EXPOSE 8082
CMD ["./gomotz"]
