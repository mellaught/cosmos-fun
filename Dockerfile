# Simple usage with a mounted data directory:
# > docker build -t onlife .
# > docker run -it -p 36657:36657 -p 36656:36656 -v ~/.onlifed:/root/.onlifed -v ~/.onlifecli:/root/.onlifecli onlife onlifed init mynode
# > docker run -it -p 36657:36657 -p 36656:36656 -v ~/.onlifed:/root/.onlifed -v ~/.onlifecli:/root/.onlifecli onlife onlifed start
FROM golang:alpine AS build-env

# Install minimum necessary dependencies, remove packages
RUN apk add --no-cache curl make git libc-dev bash gcc linux-headers eudev-dev

# Set working directory for the build
WORKDIR /go/src/github.com/mellaught/cosmos-fun

# Add source files
COPY . .

# Build onlife
RUN cd testnet/ && GOPROXY=http://goproxy.cn make install

# Final image
FROM alpine:edge

WORKDIR /root

# Copy over binaries from the build-env
COPY --from=build-env /go/bin/onlifed /usr/bin/onlifed
COPY --from=build-env /go/bin/onlifed /usr/bin/onlifecli

CMD ["onlifed"]
