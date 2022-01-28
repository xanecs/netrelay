FROM golang:alpine AS build
COPY go.mod *.go /app/
WORKDIR /app
RUN go build github.com/xanecs/netrelay

FROM ghcr.io/tailscale/tailscale:latest
COPY deploy/run.sh /run.sh
COPY --from=build /app/netrelay /netrelay
CMD "/run.sh"