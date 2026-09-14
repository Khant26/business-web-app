FROM node:24-alpine AS web
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend ./
RUN npm run build

FROM golang:1.23-alpine AS api
WORKDIR /app/backend
COPY backend/go.mod ./
COPY backend/main.go ./
RUN CGO_ENABLED=0 go build -o /northstar .

FROM alpine:3.21
WORKDIR /app
COPY --from=api /northstar ./northstar
COPY --from=web /app/frontend/dist ./frontend/dist
RUN addgroup -S app && adduser -S app -G app && mkdir -p /app/data && chown -R app:app /app
USER app
ENV PORT=8080 WEB_DIR=/app/frontend/dist DATA_FILE=/app/data/orders.json
EXPOSE 8080
CMD ["./northstar"]
