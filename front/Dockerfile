FROM node:20-alpine as build-stage

ARG MODE=production

WORKDIR /app
RUN corepack enable

COPY .npmrc package.json pnpm-lock.yaml ./
RUN --mount=type=cache,id=pnpm-store,target=/root/.pnpm-store \
    pnpm install --frozen-lockfile

COPY . .
RUN pnpm build --mode ${MODE}

EXPOSE 5000

CMD ["pnpm", "run", "server"]
