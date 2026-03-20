FROM node:20-alpine AS build
WORKDIR /app
COPY client/package*.json ./
RUN npm ci
COPY client/. .
RUN npm run build

FROM nginx:1.27-alpine
COPY --from=build /app/dist/go-lab-client /usr/share/nginx/html
EXPOSE 80
