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

EXPOSE 50051

CMD ["/tempconv-server"]
