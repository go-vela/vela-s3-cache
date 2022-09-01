
# Copyright (c) 2022 Target Brands, Inc. All rights reserved.
#
# Use of this source code is governed by the LICENSE file in this repository.

#############################################################################
##     docker build --no-cache --target certs -t vela-s3-cache:certs .     ##
#############################################################################

FROM alpine as certs

RUN apk add --update --no-cache ca-certificates

##############################################################
##     docker build --no-cache -t vela-s3-cache:local .     ##
##############################################################

FROM scratch

COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

COPY release/vela-s3-cache /bin/vela-s3-cache

ENTRYPOINT [ "/bin/vela-s3-cache" ]
