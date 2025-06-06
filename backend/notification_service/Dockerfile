FROM golang:1.23.4-alpine

WORKDIR /app

COPY . .

RUN go build -o application main.go

EXPOSE 8091

CMD ["./application"]
