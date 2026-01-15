# build stage
FROM harness0.harness.io/oci/docker_artifacts/ubi9/go1:1.24.11 AS builder
ENV GOFLAGS="-mod=readonly"

#RUN apt-get update && apt-get install -y ca-certificates make git curl mercurial
#RUN apk add --update --no-cache ca-certificates make git curl mercurial
USER root
RUN microdnf update && microdnf install -y ca-certificates make git && microdnf clean all

RUN mkdir -p /build
WORKDIR /build

COPY go.* /build/
RUN go mod download

COPY . /build
RUN BINARY_NAME=telescopes make build-release

# FROM alpine:3.14.0
FROM harness0.harness.io/oci/docker_artifacts/ubi9/go1:1.24.11
USER root
# RUN microdnf install yum
# RUN apk add --update --no-cache ca-certificates tzdata bash curl
# RUN yum install -y ca-certificates tzdata

COPY --from=builder /build/build/release/telescopes /bin

ENTRYPOINT ["/bin/telescopes"]
CMD ["/bin/telescopes"]

