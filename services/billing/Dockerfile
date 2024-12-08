FROM golang:1.23.4-alpine3.20
EXPOSE 8082
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY payment.go ./main.go
COPY .env /
RUN CGO_ENABLED=0 GOOS=linux go build -o /billing-serv
WORKDIR /
CMD ["/billing-serv"]
