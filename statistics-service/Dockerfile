FROM node:24-alpine
WORKDIR /app
RUN npm install -g pnpm@11.22.0
COPY package.json pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile --prod
COPY src ./src
COPY openapi.yaml docs.html ./
USER node
EXPOSE 3000
CMD ["node", "src/server.js"]
