FROM golang:1.26 AS builder

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o tempconv-server ./cmd/server

FROM gcr.io/distroless/static-debian12

WORKDIR /

COPY --from=builder /app/tempconv-server /tempconv-server

# Cloud Run uses PORT=8080
ENV PORT=8080
EXPOSE 8080

CMD ["/tempconv-server"]
