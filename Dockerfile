# MIT License
#
# (C) Copyright [2022] Hewlett Packard Enterprise Development LP
#
# Permission is hereby granted, free of charge, to any person obtaining a
# copy of this software and associated documentation files (the "Software"),
# to deal in the Software without restriction, including without limitation
# the rights to use, copy, modify, merge, publish, distribute, sublicense,
# and/or sell copies of the Software, and to permit persons to whom the
# Software is furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included
# in all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
# THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
# OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
# ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
# OTHER DEALINGS IN THE SOFTWARE.

# Dockerfile for building HMS Discovery Service.


### Build Base Stage ###
# Build base just has the packages installed we need.
FROM arti.dev.cray.com/baseos-docker-master-local/golang:1.16-alpine3.13 AS build-base

RUN set -ex \
    && apk -U upgrade \
    && apk add build-base


### Base Stage ###
# Base copies in the files we need to test/build.
FROM build-base AS base

RUN go env -w GO111MODULE=auto

# Copy all the necessary files to the image.
COPY cmd $GOPATH/src/github.com/Cray-HPE/hms-discovery-service/cmd
COPY internal $GOPATH/src/github.com/Cray-HPE/hms-discovery-service/internal
COPY pkg $GOPATH/src/github.com/Cray-HPE/hms-discovery-service/pkg
COPY vendor $GOPATH/src/github.com/Cray-HPE/hms-discovery-service/vendor


### Build Stage ###
FROM base AS builder

# Base image contains everything needed for Go building, just build.
RUN set -ex \
    && go build -v -o /usr/local/bin/hmds -i github.com/Cray-HPE/hms-discovery-service/cmd/discovery-service


### Final Stage ###
FROM arti.dev.cray.com/baseos-docker-master-local/alpine:3.13
LABEL maintainer="Hewlett Packard Enterprise" 
EXPOSE 27779
STOPSIGNAL SIGTERM

# Copy the final binary
COPY --from=builder /usr/local/bin/hmds /usr/local/bin

# Cannot live without these packages installed.
RUN set -ex \
    && apk -U upgrade \
    && apk add --no-cache \
        curl

# nobody 65534:65534
USER 65534:65534

CMD ["sh", "-c", "hmds"]
