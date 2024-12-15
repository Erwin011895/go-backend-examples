FROM golang:alpine AS builder

WORKDIR /myapp

COPY . .

RUN go build -o app .

EXPOSE 8081

CMD ["./app"]

# FROM alpine

# WORKDIR /app

# COPY --from=builder /app/myapp .

# EXPOSE 8081

# CMD ["./app"]
