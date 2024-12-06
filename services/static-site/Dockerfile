FROM golang:1.23.4-alpine3.20
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY main.go ./
COPY static/ /static/ 
RUN CGO_ENABLED=0 GOOS=linux go build -o /static-serv
WORKDIR /
CMD ["/static-serv"]
