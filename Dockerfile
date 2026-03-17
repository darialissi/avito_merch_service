FROM golang:1.25-alpine AS builder

WORKDIR /build

# Кэширование зависимостей
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -mod=readonly -o /build/server ./cmd

FROM scratch

WORKDIR /app

COPY --from=builder /build/server ./server
COPY --from=builder /build/configs ./configs

CMD ["./server"]