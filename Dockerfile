FROM golang:1.26-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go build -o server .


FROM alpine:latest

WORKDIR /app

COPY --from=build /app/server .
COPY --from=build /app/templates ./templates

EXPOSE 8080
CMD ["./server"]