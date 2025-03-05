FROM golang

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .
RUN CGO_ENABLED=0 go build -o main ./cmd/bot/main.go

CMD ["./main -config ./configs/config.yaml"]