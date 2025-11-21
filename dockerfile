# templ + go build
FROM golang:1.24 AS build-stage
WORKDIR /app

# install templ
RUN go install github.com/a-h/templ/cmd/templ@latest

# copy go.mod early for caching
COPY go.mod go.sum ./
RUN go mod download

# copy source
COPY . .

# generate .go files from templ
RUN templ generate

# build Go binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /entrypoint


# tailwind build
FROM node:20-alpine AS tailwind-stage
WORKDIR /app

# copy configs
COPY tailwind.config.js package.json package-lock.json ./

# copy assets
COPY assets ./assets

# copy templ files so tailwind can scan them
COPY views ./views

RUN npm ci
RUN mkdir -p ./assets/css
RUN npx tailwindcss -i ./assets/css/input.css -o ./assets/css/style.css --minify


# final release stage
FROM gcr.io/distroless/static-debian11 AS release-stage
WORKDIR /

COPY --from=build-stage /entrypoint /entrypoint
COPY --from=tailwind-stage /app/assets /assets

EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/entrypoint"]
