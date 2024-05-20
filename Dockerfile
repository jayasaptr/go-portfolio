FROM golang:1.22-alpine

WORKDIR /app

COPY . .

RUN go build -o portfolio-api

EXPOSE 8080

CMD ./portfolio-api