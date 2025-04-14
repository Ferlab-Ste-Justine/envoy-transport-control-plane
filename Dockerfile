FROM golang:1.23-bullseye as builder

ENV CGO_ENABLED=0

WORKDIR /opt
COPY . .

RUN go build .

FROM scratch

COPY --from=builder /opt/envoy-transport-control-plane /bin/

ENV WORKING_DIR="/opt"

ENTRYPOINT ["/bin/envoy-transport-control-plane"]