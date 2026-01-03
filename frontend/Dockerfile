# Build stage
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
RUN npm run build

# Production stage
FROM alpine:latest
RUN apk add --no-cache caddy
WORKDIR /app
COPY --from=builder /app/dist /usr/share/caddy
EXPOSE 3000
CMD ["caddy", "file-server", "--listen", ":3000", "--root", "/usr/share/caddy"]
