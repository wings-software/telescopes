# Builder: hhios/go1 sets GOFIPS140. 1.27-fips has make/go but not git/apk.
FROM harness0.harness.io/oci/docker_artifacts/hhios/go1:1.27-fips AS builder
ENV GOFLAGS="-mod=readonly"

USER root

RUN mkdir -p /build
WORKDIR /build

# CI Build_Services passes these; Makefile uses them for ldflags when git is absent.
ARG BUILD_VERSION
ARG COMMIT_SHA
ARG BRANCH_NAME
ENV VERSION=${BUILD_VERSION} COMMIT_HASH=${COMMIT_SHA} BRANCH=${BRANCH_NAME}

COPY go.* /build/
RUN go mod download

COPY . /build
RUN BINARY_NAME=telescopes make build-release

# Runtime: image ENTRYPOINT is fips-entrypoint.sh (honours FIPS_ENABLED). Do not copy it or set ENTRYPOINT.
FROM harness0.harness.io/oci/docker_artifacts/hhios/go1:1.27-fips

COPY --from=builder /build/build/release/telescopes /bin

CMD ["/bin/telescopes"]
