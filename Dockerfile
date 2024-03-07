FROM --platform=${BUILDPLATFORM} golang:1-alpine as builder
ARG TARGETOS
ARG TARGETARCH
ENV CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH}

RUN apk add --no-cache git

COPY go.mod go.sum /app/
WORKDIR /app
RUN go mod download

COPY . /app

RUN --mount=type=cache,target=/root/go \
  go install github.com/a-h/templ/cmd/templ@latest && \
  go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

RUN sqlc generate && templ generate

RUN --mount=type=cache,target=/root/.cache/go-build \
  go build -ldflags '-s -w' -trimpath -v -o /bin/app

FROM alpine:3

RUN apk upgrade --no-cache && \
  apk add --no-cache tini bash ca-certificates curl
RUN addgroup -S serializer && \
  adduser -S -G serializer serializer
RUN mkdir -p /app/data /app/bin && \
  chown -R serializer:serializer /app/data

ENV GO_ENV=production

EXPOSE 3000
WORKDIR /app

COPY --from=builder /bin/app /app/app

USER serializer
WORKDIR /app
HEALTHCHECK --interval=30s --timeout=3s CMD ["curl", "-sqf", "http://localhost:3000/healthz"]
CMD ["/sbin/tini", "--", "/app/app"]
