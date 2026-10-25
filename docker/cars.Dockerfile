FROM golang:1.24.2-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
COPY migrations ./migrations
COPY templates ./templates
COPY resources ./resources
RUN CGO_ENABLED=0 go build -trimpath -o /out/cars ./cmd/cars && \
    CGO_ENABLED=0 go build -trimpath -o /out/cars-migrate ./cmd/migrate

FROM debian:bookworm-slim
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
WORKDIR /app
COPY --from=build /out/ /usr/local/bin/
COPY templates ./templates
COPY resources ./resources
USER 65534:65534
EXPOSE 8082
CMD ["cars"]
