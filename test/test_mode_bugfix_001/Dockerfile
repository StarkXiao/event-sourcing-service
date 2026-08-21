FROM golang:1.26

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build ./...

EXPOSE 8080
ENV HTTP_ADDR=:8080 AUTO_MIGRATE=true
CMD ["go", "run", "./cmd/server"]
