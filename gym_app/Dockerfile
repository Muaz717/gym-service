FROM golang:1.24-alpine3.21 AS builder

WORKDIR /usr/local/src

RUN apk --no-cache add bash git make gcc gettext musl-dev

# Dependencies
COPY ["app/go.mod", "app/go.sum", "./"]
RUN go mod download

# Build
COPY app ./
RUN go build -o ./bin/gym_app  cmd/gym-app/main.go

FROM alpine:3.21 AS runner

COPY --from=builder usr/local/src/bin/gym_app /
COPY config/prod.yaml /config/prod.yaml

CMD ["/gym_app"]
