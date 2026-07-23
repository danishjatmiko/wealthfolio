FROM node:20-alpine AS build
WORKDIR /src
ARG VITE_API_BASE_URL=/api/v1
ENV VITE_API_BASE_URL=$VITE_API_BASE_URL
COPY package.json package-lock.json* ./
RUN npm ci
COPY . .
RUN npm run build

FROM caddy:2-alpine
RUN addgroup -S caddy && adduser -S caddy -G caddy && \
    mkdir -p /data/caddy /config/caddy && \
    chown -R caddy:caddy /data /config
COPY --from=build /src/dist /srv
COPY Caddyfile /etc/caddy/Caddyfile
USER caddy
EXPOSE 80 443
