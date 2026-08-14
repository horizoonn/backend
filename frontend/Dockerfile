# Сборка статики
FROM node:22-slim AS build

WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci --no-audit --no-fund

COPY . .

# Год итогов вшивается в бандл на сборке: vite подставляет import.meta.env.
ARG VITE_RECAP_YEAR
ENV VITE_RECAP_YEAR=$VITE_RECAP_YEAR

RUN npm run build

# Раздача через nginx: он же проксирует /api на Go-сервис,
# поэтому браузер видит один origin и CORS не нужен.
FROM nginx:1.27-alpine

COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist /usr/share/nginx/html

EXPOSE 80
