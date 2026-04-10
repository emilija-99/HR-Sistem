FROM golang:1.25-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o server ./cmd/api
RUN chmod +x /app/server

EXPOSE 8080
CMD ["./server"]
