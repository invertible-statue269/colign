FROM node:20-alpine

WORKDIR /app

COPY package.json ./
RUN npm install --production

COPY src/ ./src/

EXPOSE 1234

CMD ["node", "--import", "tsx", "src/index.ts"]
