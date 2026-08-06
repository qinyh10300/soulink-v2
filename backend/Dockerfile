FROM golang:1.26-alpine AS builder
WORKDIR /src
ARG GOPROXY=https://proxy.golang.org,direct
ENV GOPROXY=${GOPROXY}

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /out/soulink-api ./cmd/api

FROM alpine:3.22
WORKDIR /app
RUN apk add --no-cache ca-certificates python3 tzdata wget

COPY --from=builder /out/soulink-api /app/soulink-api
COPY --from=builder /src/configs/match /app/configs/match
COPY --from=builder /src/soulagent /app/soulagent

EXPOSE 8080
ENTRYPOINT ["/app/soulink-api"]
