FROM golang:1.24-alpine as builder

WORKDIR /app

COPY go.* ./

RUN go mod download

COPY . .

RUN go build -o main cmd/server/main.go

EXPOSE 8090

CMD [ "./main" ]
