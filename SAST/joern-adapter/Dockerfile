FROM golang:1.23-bookworm AS build
WORKDIR /src
COPY go.mod ./
COPY cmd cmd
COPY internal internal
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/sast-server ./cmd/adapter

FROM eclipse-temurin:17-jdk-jammy
RUN apt-get update \
    && apt-get install -y --no-install-recommends bash curl \
    && rm -rf /var/lib/apt/lists/* \
    && groupadd --system app \
    && useradd --system --gid app --home-dir /app app
WORKDIR /app
COPY --from=build /out/sast-server /app/sast-server
USER app
EXPOSE 8087
ENTRYPOINT ["/app/sast-server"]
