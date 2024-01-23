FROM golang:1.21 as builder

COPY . /opt/project/
WORKDIR /opt/project

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o ./deployer

FROM alpine:3.19.0
COPY --from=builder /opt/project/deployer /deployer
ENTRYPOINT ["/deployer"]