FROM golang:latest
LABEL maintainer="Enrique Berrueta <eabz@polispay.org>"

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .
EXPOSE 8080
CMD ["./main"]