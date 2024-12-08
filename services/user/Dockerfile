FROM golang:1.23.4-alpine3.20
EXPOSE 80
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY main.go ./main.go
COPY .env /
RUN CGO_ENABLED=0 GOOS=linux go build -o /user-serv
WORKDIR /
CMD ["/user-serv"]
