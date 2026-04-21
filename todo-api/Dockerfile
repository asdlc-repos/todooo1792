FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/todo-api ./

FROM gcr.io/distroless/static:nonroot

WORKDIR /app

COPY --from=builder /out/todo-api /app/todo-api

USER nonroot:nonroot

EXPOSE 9090

ENTRYPOINT ["/app/todo-api"]
