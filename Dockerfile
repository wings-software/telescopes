# build stage
# hhi/go-fips replaces ubi9/go1 *-fips-rfcurated: Go 1.26.10 built with the Go FIPS 140-3 module
# (GOFIPS140 is set by the image). The -dev variant carries git, make, curl and apk.
FROM harness0.harness.io/oci/docker_artifacts/hhi/go-fips:1.26-dev AS builder
ENV GOFLAGS="-mod=readonly"

USER root
# git, make, curl and CA certificates ship in go-fips:1.26-dev (no microdnf in this image).
RUN git --version && make --version | head -1

RUN mkdir -p /build
WORKDIR /build

COPY go.* /build/
RUN go mod download

COPY . /build
RUN BINARY_NAME=telescopes make build-release

# runtime: hhi/go-fips (Go FIPS 140-3 + OpenSSL FIPS provider). It sets no ENTRYPOINT/CMD of its own.
FROM harness0.harness.io/oci/docker_artifacts/hhi/go-fips:1.26
USER root

COPY --from=builder /build/build/release/telescopes /bin
# FIPS_ENABLED=false switches FIPS off at run time; otherwise the image runs exactly as shipped.
COPY --chmod=755 fips-entrypoint.sh /usr/local/bin/fips-entrypoint.sh

ENTRYPOINT ["/usr/local/bin/fips-entrypoint.sh", "/bin/telescopes"]
# CMD kept identical to the previous image (the app receives it as an argument).
CMD ["/bin/telescopes"]
