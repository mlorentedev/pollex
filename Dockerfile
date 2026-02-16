FROM golang:1.26-alpine AS builder

ARG VERSION=dev
ARG VCS_REF=unknown

WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .

RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=${VERSION}" -o /pollex ./cmd/pollex

FROM alpine:3.21

ARG VERSION=dev
ARG VCS_REF=unknown

RUN apk add --no-cache ca-certificates curl \
    && addgroup -g 1001 -S pollex \
    && adduser -S pollex -u 1001 -G pollex

COPY --from=builder /pollex /pollex
COPY prompts/polish.txt /etc/pollex/polish.txt

USER pollex:pollex

ENV POLLEX_PORT=8090
ENV POLLEX_PROMPT_PATH=/etc/pollex/polish.txt
EXPOSE 8090

LABEL org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${VCS_REF}" \
      org.opencontainers.image.source="https://github.com/mlorentedev/pollex"

ENTRYPOINT ["/pollex"]
