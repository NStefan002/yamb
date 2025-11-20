# build stage
FROM golang:1.24 AS build-stage
WORKDIR /app

# install templ
RUN go install github.com/a-h/templ/cmd/templ@latest

# Install Tailwind CLI
RUN apt-get update && apt-get install -y curl && \
    curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64 && \
    chmod +x tailwindcss-linux-x64 && mv tailwindcss-linux-x64 /usr/local/bin/tailwindcss

# copy go.mod first for caching
COPY go.mod go.sum ./
RUN go mod download

# copy everything
COPY . /app

# generate go files from templ files
RUN templ generate

# generate tailwind css
RUN tailwindcss -i ./assets/css/input.css -o ./assets/css/style.css

# build Go binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /entrypoint

# release stage
FROM gcr.io/distroless/static-debian11 AS release-stage
WORKDIR /

COPY --from=build-stage /entrypoint /entrypoint
COPY --from=build-stage /app/assets /assets

EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/entrypoint"]
