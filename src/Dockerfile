FROM golang:1.23.1-alpine

WORKDIR /app

COPY . .

RUN go build -o node-app .

CMD ["./node-app"]
