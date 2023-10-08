FROM golang:1.20.8-alpine

WORKDIR /app

COPY . .

RUN go build -o build/

EXPOSE 8080
CMD ["./build/docker-intro"]