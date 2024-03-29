
# SPDX-License-Identifier: Apache-2.0

#############################################################################
##     docker build --no-cache --target certs -t vela-s3-cache:certs .     ##
#############################################################################

FROM alpine:3.19@sha256:c5b1261d6d3e43071626931fc004f70149baeba2c8ec672bd4f27761f8e1ad6b as certs

RUN apk add --update --no-cache ca-certificates

##############################################################
##     docker build --no-cache -t vela-s3-cache:local .     ##
##############################################################

FROM scratch

COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

COPY release/vela-s3-cache /bin/vela-s3-cache

ENTRYPOINT [ "/bin/vela-s3-cache" ]
