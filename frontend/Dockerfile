FROM golang:1.26-alpine AS build

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN go build -o frontend-server .

FROM alpine:latest

WORKDIR /app

COPY --from=build /app/frontend-server .
COPY --from=build /app/index.html .

EXPOSE 8080

CMD ["./frontend-server"]