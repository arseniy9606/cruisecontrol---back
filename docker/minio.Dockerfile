# Official sources, pinned to preserve the console used by the Lab 1 tutorial.
FROM golang:1.24.2-bookworm AS minio-build
ENV CGO_ENABLED=0
RUN go install github.com/minio/minio@RELEASE.2025-04-22T22-12-26Z

FROM debian:bookworm-slim AS minio
COPY --from=minio-build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=minio-build /go/bin/minio /usr/local/bin/minio
LABEL org.opencontainers.image.source="https://github.com/minio/minio" \
      org.opencontainers.image.version="RELEASE.2025-04-22T22-12-26Z" \
      org.opencontainers.image.licenses="AGPL-3.0-only"
EXPOSE 9000 9001
ENTRYPOINT ["minio"]

FROM golang:1.24.2-bookworm AS mc-build
ENV CGO_ENABLED=0
RUN go install github.com/minio/mc@RELEASE.2025-04-16T18-13-26Z

FROM debian:bookworm-slim AS media-init
COPY --from=mc-build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=mc-build /go/bin/mc /usr/local/bin/mc
LABEL org.opencontainers.image.source="https://github.com/minio/mc" \
      org.opencontainers.image.licenses="AGPL-3.0-only"
ENTRYPOINT ["/bin/sh"]
