# API-only image (frontend hosted separately, e.g. on Vercel).
# Build from the backend folder:  docker build -t sawyun-api ./backend
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/sawyun ./cmd/api \
 && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/sawyun-seed ./cmd/seed

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata wget && addgroup -S app && adduser -S app -G app
WORKDIR /app
COPY --from=build /out/sawyun /out/sawyun-seed ./
RUN mkdir -p /app/uploads && chown -R app:app /app
USER app
ENV APP_ENV=production PORT=8080 UPLOAD_DIR=/app/uploads WEB_DIR=/app/web TZ=Asia/Bangkok
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s --retries=3 CMD wget -qO- http://127.0.0.1:8080/api/v1/health || exit 1
CMD ["./sawyun"]
