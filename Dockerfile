FROM --platform=${BUILDPLATFORM} golang:1 AS builder
ARG TARGETOS
ARG TARGETARCH
ENV GOOS=${TARGETOS} GOARCH=${TARGETARCH}

RUN apt update && \
  apt install -y git

RUN --mount=type=cache,target=/root/go \
  --mount=type=cache,target=/root/.cache/go-build \
  go install github.com/a-h/templ/cmd/templ@latest && \
  go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

COPY go.mod go.sum /app/
WORKDIR /app

RUN --mount=type=cache,target=/root/go \
  go mod download

COPY . /app

RUN --mount=type=cache,target=/root/go \
  --mount=type=cache,target=/root/.cache/go-build \
  sqlc generate && \
  templ generate

RUN --mount=type=cache,target=/root/go \
  --mount=type=cache,target=/root/.cache/go-build \
  go build -ldflags '-s -w' -trimpath -v -o /bin/app

FROM debian:12-slim

RUN apt update && \
  apt install -y tini curl
RUN groupadd --system -g 105 serializer && \
  useradd --system -d /home/serializer -m -u 105 -g serializer serializer
RUN mkdir -p /app/data /app/bin && \
  chown -R serializer:serializer /app/data

ENV GO_ENV=production

EXPOSE 3000
WORKDIR /app

COPY --from=builder /bin/app /app/app

USER serializer
RUN id
WORKDIR /app
HEALTHCHECK --interval=30s --timeout=3s CMD ["curl", "-sqf", "http://localhost:3000/healthz"]
CMD ["/usr/bin/tini", "--", "/app/app"]
