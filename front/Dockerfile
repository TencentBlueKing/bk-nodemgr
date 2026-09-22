FROM node:20-alpine AS build-stage

ARG MODE=production
# 构建镜像用不到 cypress，跳过 postinstall 下载 ~250MB 二进制
ENV CYPRESS_INSTALL_BINARY=0

WORKDIR /app
RUN corepack enable

COPY .npmrc package.json pnpm-lock.yaml ./
# 显式把 store 指到 cache mount（pnpm 9 默认 store 在 ~/.local/share/pnpm/store，
# 不配置的话下面的 cache mount 不会命中，等于每次全量下载依赖）
RUN --mount=type=cache,id=pnpm-store,target=/root/.pnpm-store \
    pnpm config set store-dir /root/.pnpm-store --global && \
    pnpm install --frozen-lockfile

COPY . .
RUN pnpm build --mode ${MODE}

# 运行镜像只带构建产物，不含全部 node_modules（~1GB+ -> ~250MB）
FROM node:20-alpine
WORKDIR /app
COPY --from=build-stage /app/dist ./dist
RUN npm install -g http-server@14.1.1
EXPOSE 5000
CMD ["http-server", "./dist", "-p", "5000"]
