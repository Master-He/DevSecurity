FROM golang:1.23-bookworm AS build
WORKDIR /src
COPY go.mod ./
COPY cmd cmd
COPY internal internal
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/sca ./cmd/sca
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/sca-server ./cmd/sca-server

FROM maven:3.9.9-eclipse-temurin-17
RUN apt-get update \
    && apt-get install -y --no-install-recommends curl \
    && rm -rf /var/lib/apt/lists/* \
    && groupadd --system --gid 10001 app \
    && useradd --system --uid 10001 --gid app --home-dir /app app \
    && mkdir -p /var/lib/devsecurity/sca-work \
    && chown -R app:app /var/lib/devsecurity
WORKDIR /app
COPY --from=build /out/sca /app/sca
COPY --from=build /out/sca-server /app/sca-server
COPY data /app/data
USER app
EXPOSE 8086
ENTRYPOINT ["/app/sca-server"]
