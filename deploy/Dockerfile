FROM golang:alpine AS build
COPY go.mod *.go /app/
ENV CGO_ENABLED=0
WORKDIR /app
RUN go build github.com/xanecs/netrelay

FROM scratch
COPY --from=build /app/netrelay /netrelay
ENTRYPOINT [ "/netrelay" ]