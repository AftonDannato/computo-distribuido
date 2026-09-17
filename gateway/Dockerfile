FROM golang:1.26-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go build -o gateway .


FROM alpine:latest

WORKDIR /app

COPY --from=build /app/gateway .
COPY --from=build /app/routes.json .

EXPOSE 8080
CMD ["./gateway"]