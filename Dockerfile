# build stage
# hhios/go1:1.26-fips: Go 1.26 with the Go FIPS 140-3 module (GOFIPS140 is set by the image) and the
# build tools this stage needs (git, make, curl, apk), so there is no package install step.
FROM harness0.harness.io/oci/docker_artifacts/hhios/go1:1.26-fips AS builder
ENV GOFLAGS="-mod=readonly"

USER root
RUN git --version && make --version | head -1

RUN mkdir -p /build
WORKDIR /build

COPY go.* /build/
RUN go mod download

COPY . /build
RUN BINARY_NAME=telescopes make build-release

# runtime: hhios/go1:1.26-fips. The image already has ENTRYPOINT fips-entrypoint.sh, which honours
# FIPS_ENABLED and then runs CMD, so this Dockerfile must not copy that script or set an ENTRYPOINT.
FROM harness0.harness.io/oci/docker_artifacts/hhios/go1:1.26-fips

COPY --from=builder /build/build/release/telescopes /bin

CMD ["/bin/telescopes"]
