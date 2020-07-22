FROM golang:1.13

WORKDIR /app
COPY . .

RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN go build -o main /app/main.go

EXPOSE 80

CMD ["/app/main", "--port=80"]