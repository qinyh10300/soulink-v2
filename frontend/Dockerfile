# ── Stage 1: Build ────────────────────────────────
FROM node:22-alpine AS builder

WORKDIR /app

ARG VITE_DEMO_MODE=false
ARG VITE_DEMO_EMAIL=""
ARG VITE_DEMO_PASSWORD=""
ARG VITE_DEMO_ADMIN_EMAIL=""
ARG VITE_DEMO_ADMIN_PASSWORD=""
ARG NPM_REGISTRY=https://registry.npmjs.org
ARG NPM_FALLBACK_REGISTRY=https://registry.npmjs.org
ENV VITE_DEMO_MODE=${VITE_DEMO_MODE}
ENV VITE_DEMO_EMAIL=${VITE_DEMO_EMAIL}
ENV VITE_DEMO_PASSWORD=${VITE_DEMO_PASSWORD}
ENV VITE_DEMO_ADMIN_EMAIL=${VITE_DEMO_ADMIN_EMAIL}
ENV VITE_DEMO_ADMIN_PASSWORD=${VITE_DEMO_ADMIN_PASSWORD}

COPY package.json package-lock.json ./
RUN --mount=type=cache,target=/root/.npm \
    npm ci \
      --registry="${NPM_REGISTRY}" \
      --replace-registry-host=always \
      --fetch-retries=0 \
      --fetch-timeout=120000 \
      --no-audit \
      --no-fund || \
    (rm -rf node_modules && \
      npm ci \
        --registry="${NPM_FALLBACK_REGISTRY}" \
        --replace-registry-host=always \
        --fetch-retries=2 \
        --fetch-timeout=300000 \
        --no-audit \
        --no-fund)

COPY . .
RUN npm run build

# ── Stage 2: Serve ────────────────────────────────
FROM nginx:alpine

ARG NGINX_CONF=nginx.conf
COPY ${NGINX_CONF} /etc/nginx/conf.d/default.conf
COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80 443

CMD ["nginx", "-g", "daemon off;"]
