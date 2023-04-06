# build stage
#FROM golang:1.19.2-buster AS builder
#ENV GOFLAGS="-mod=readonly"
#
#RUN apt-get update && apt-get install -y ca-certificates make git curl mercurial
#
#RUN mkdir -p /build
#WORKDIR /build
#
#COPY go.* /build/
#RUN go mod download
#
#COPY . /build
#RUN BINARY_NAME=telescopes make build-release
#
#FROM us.gcr.io/platform-205701/ubi/ubi-go:latest
#USER root
#
#COPY --from=builder /build/build/release/telescopes /bin
#
#ENTRYPOINT ["/bin/telescopes"]
#CMD ["/bin/telescopes"]
#
FROM us.gcr.io/platform-205701/ubi/ubi-go:latest AS builder
ENV GOFLAGS="-mod=readonly"

USER root
RUN microdnf update && microdnf install -y ca-certificates make git curl && microdnf clean all
USER default

RUN mkdir -p /build
WORKDIR /build

COPY go.* /build/
RUN go mod download

COPY . /build
RUN BINARY_NAME=telescopes make build-release

FROM us.gcr.io/platform-205701/ubi/ubi-go:latest
USER root

COPY --from=builder /build/build/release/telescopes /bin

ENTRYPOINT ["/bin/telescopes"]
CMD ["/bin/telescopes"]


