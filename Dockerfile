FROM golang:1.25-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
ARG VERSION=dev
RUN CGO_ENABLED=0 go build -ldflags="-s -w -X 'github.com/lakshmanpatel/tok/internal/commands/shared.Version=${VERSION}'" -o /tok ./cmd/tok/

FROM alpine:3.21
RUN apk add --no-cache ca-certificates git
COPY --from=builder /tok /usr/local/bin/tok

ENTRYPOINT ["tok"]
CMD ["--help"]
