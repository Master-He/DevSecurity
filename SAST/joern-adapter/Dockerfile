FROM golang:1.23-bookworm AS build
WORKDIR /src
COPY go.mod ./
COPY cmd cmd
COPY internal internal
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/joern-adapter ./cmd/adapter

FROM ghcr.io/joernio/joern:nightly
COPY --from=build /out/joern-adapter /usr/local/bin/joern-adapter
EXPOSE 8087
ENTRYPOINT ["/usr/local/bin/joern-adapter"]
