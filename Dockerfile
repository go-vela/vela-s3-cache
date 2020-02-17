
# Copyright (c) 2020 Target Brands, Inc. All rights reserved.
#
# Use of this source code is governed by the LICENSE file in this repository.

#############################################################################
##    docker build --no-cache --target certs -t vela-downstream:certs .    ##
#############################################################################

FROM alpine as certs

RUN apk add --update --no-cache ca-certificates

##############################################################
##    docker build --no-cache -t vela-downstream:local .    ##
##############################################################

FROM scratch

COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

COPY release/s3-cache-plugin /bin/s3-cache-plugin

ENTRYPOINT [ "/bin/s3-cache-plugin" ]