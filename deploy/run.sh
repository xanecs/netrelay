#! /bin/sh

export PATH=$PATH:/tailscale/bin

AUTH_KEY="${AUTH_KEY:-}"
KUBE_SECRET="${KUBE_SECRET:-tailscale}"

set -e

TAILSCALED_ARGS="--state=kube:${KUBE_SECRET} --socket=/tmp/tailscaled.sock --tun=userspace-networking"
UP_ARGS="--accept-dns=false --authkey=${AUTH_KEY}"

echo "Starting tailscaled"
tailscaled ${TAILSCALED_ARGS} &
PID=$!

echo "Running tailscale up"
tailscale --socket=/tmp/tailscaled.sock up ${UP_ARGS}

/netrelay
wait ${PID}
