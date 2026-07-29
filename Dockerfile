FROM golang:1.25-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/recipe-api ./cmd

FROM alpine:3.22

RUN apk add --no-cache ca-certificates \
	&& addgroup -S app \
	&& adduser -S app -G app \
	&& mkdir -p /data/images \
	&& chown -R app:app /data

WORKDIR /app

COPY --from=build /out/recipe-api /app/recipe-api

USER app

EXPOSE 8080

ENTRYPOINT ["/app/recipe-api"]
