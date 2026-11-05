# build stage
FROM harness0.harness.io/oci/docker_artifacts/hhi/go-fips:1.26-dev AS builder
ENV GOFLAGS="-mod=readonly"

USER root

RUN mkdir -p /build
WORKDIR /build

COPY go.* /build/
RUN go mod download

COPY . /build
RUN BINARY_NAME=telescopes make build-release

# runtime
FROM harness0.harness.io/oci/docker_artifacts/hhi/go-fips:1.26
USER root

COPY --from=builder /build/build/release/telescopes /bin

ENTRYPOINT ["/bin/telescopes"]
CMD ["/bin/telescopes"]
