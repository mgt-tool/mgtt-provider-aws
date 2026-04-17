# Multi-stage build for mgtt-provider-aws.
# The provider binary shells out to `aws` (AWS CLI v2), so the runtime
# image is the official amazon/aws-cli base with our provider binary
# grafted in. That image ships a glibc userland; statically-linked
# Go binaries drop in cleanly.
#
# The base image's default ENTRYPOINT is `aws`; we override with the
# provider binary and keep aws reachable on PATH.

FROM golang:1.25 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/provider .

FROM amazon/aws-cli:2.17.0
COPY --from=build /out/provider /usr/local/bin/provider
COPY provider.yaml /provider.yaml
COPY types /types
ENTRYPOINT ["/usr/local/bin/provider"]
