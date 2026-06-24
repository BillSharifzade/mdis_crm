# ---- Build stage ----
FROM node:20-alpine AS builder

WORKDIR /app

COPY package.json package-lock.json* ./
RUN npm ci || npm install

COPY . .

# VITE_API_BASE_URL подставляется во время сборки.
# По умолчанию указывает на backend, доступный с хоста (localhost:8082).
ARG VITE_API_BASE_URL=http://localhost:8082/api/v1
ARG VITE_USE_API=true
# VITE_BASE_PATH — публичный суб-путь, под которым раздаётся SPA (например /mdis_crm/).
# По умолчанию "/" — раздача из корня домена.
ARG VITE_BASE_PATH=/
ENV VITE_API_BASE_URL=$VITE_API_BASE_URL
ENV VITE_USE_API=$VITE_USE_API
ENV VITE_BASE_PATH=$VITE_BASE_PATH

RUN npm run build

# ---- Runtime stage ----
FROM nginx:1.27-alpine

COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
