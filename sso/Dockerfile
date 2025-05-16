FROM golang:1.24-alpine3.21 AS builder

WORKDIR /usr/local/src

RUN apk --no-cache add bash git make gcc gettext musl-dev curl

# Dependencies
COPY ["app/go.mod", "app/go.sum", "./"]
RUN go mod download

# Build app
COPY app ./
RUN go build -o ./bin/sso_app cmd/sso/main.go

## Download grpc_health_probe binary
#RUN curl -sSL -o ./bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/v0.4.15/grpc_health_probe-linux-amd64 && \
#    chmod +x ./bin/grpc_health_probe

# ---- Runner ----
FROM alpine:3.21 AS runner

# Copy app binary
COPY --from=builder /usr/local/src/bin/sso_app /
#COPY --from=builder /usr/local/src/bin/grpc_health_probe /usr/local/bin/grpc_health_probe

# Copy config
COPY config/prod.yaml /config/prod.yaml

CMD ["/sso_app"]