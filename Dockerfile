# Use go 1.24 image
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY ./app .
RUN go mod init url_shortener
# Fetch dependencies
RUN go get github.com/gorilla/mux
RUN CGO_ENABLED=0 GOOS=linux go build -o server main.go

# Use the minimal image to reduce size
FROM scratch
WORKDIR /app
COPY --from=builder /app/server /app/
EXPOSE 8080
ENTRYPOINT ["./server"]