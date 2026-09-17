FROM golang:1.26-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go build -o events-backend .


FROM alpine:latest

WORKDIR /app

COPY --from=build /app/events-backend .

EXPOSE 8080
CMD ["./events-backend"]