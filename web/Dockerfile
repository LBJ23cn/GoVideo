#
# 前端生产镜像：Node 构建 Vue (Vite) 产物，交由 Nginx 托管并反向代理 API。
#
# 构建（在 web 目录下）:
#   docker build -t govideo-web .
# 通常直接由根目录 docker-compose.yml 构建，无需手动执行。
#

FROM node:24-alpine AS build
WORKDIR /src

COPY package.json package-lock.json ./
RUN npm ci

COPY . ./
RUN npm run build

FROM nginx:1.27-alpine
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /src/dist /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
