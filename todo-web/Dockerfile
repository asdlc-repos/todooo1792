# syntax=docker/dockerfile:1.6

# ---------- Build stage ----------
FROM node:20-alpine AS build
WORKDIR /app

COPY package.json ./
RUN npm install --no-audit --no-fund --loglevel=error

COPY tsconfig.json tsconfig.node.json vite.config.ts index.html ./
COPY src ./src

RUN npm run build

# ---------- Runtime stage ----------
FROM nginx:1.27-alpine AS runtime

# Static assets — index.html already references /config.js, which nginx
# renders dynamically from env vars (see nginx.conf location = /config.js).
COPY --from=build /app/dist /usr/share/nginx/html

# Nginx template — the base image's entrypoint renders any file under
# /etc/nginx/templates/*.template into /etc/nginx/conf.d/ using envsubst.
COPY nginx.conf /etc/nginx/templates/default.conf.template

# Default env values so the container runs without any configuration.
# NGINX_ENVSUBST_FILTER restricts substitution to these two vars so nginx's
# built-in $host/$remote_addr/etc. are left untouched.
ENV API_BASE_URL=/api \
    API_UPSTREAM=http://todo-api:9090 \
    NGINX_ENVSUBST_TEMPLATE_SUFFIX=.template \
    NGINX_ENVSUBST_OUTPUT_DIR=/etc/nginx/conf.d \
    NGINX_ENVSUBST_FILTER="^(API_BASE_URL|API_UPSTREAM)$"

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -qO- http://localhost:8080/health || exit 1

CMD ["nginx", "-g", "daemon off;"]
