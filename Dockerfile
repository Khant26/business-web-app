FROM node:24-alpine AS web
WORKDIR /src/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend ./
RUN npm run build

FROM golang:1.26-alpine AS api
WORKDIR /src/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/sawyun ./cmd/api

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata && addgroup -S app && adduser -S app -G app
WORKDIR /app
COPY --from=api /out/sawyun ./sawyun
COPY --from=web /src/frontend/dist ./web
RUN mkdir -p /app/uploads && chown -R app:app /app
USER app
ENV PORT=8080 WEB_DIR=/app/web UPLOAD_DIR=/app/uploads TZ=Asia/Bangkok
EXPOSE 8080
CMD ["./sawyun"]
