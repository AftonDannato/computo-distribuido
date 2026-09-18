FROM golang:1.26-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go build -o detection-backend .


FROM alpine:latest

WORKDIR /app

COPY --from=build /app/detection-backend .

EXPOSE 8080
CMD ["./detection-backend"]