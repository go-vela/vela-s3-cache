
# SPDX-License-Identifier: Apache-2.0

#############################################################################
##     docker build --no-cache --target certs -t vela-s3-cache:certs .     ##
#############################################################################

FROM alpine:3.20@sha256:0a4eaa0eecf5f8c050e5bba433f58c052be7587ee8af3e8b3910ef9ab5fbe9f5 as certs

RUN apk add --update --no-cache ca-certificates

##############################################################
##     docker build --no-cache -t vela-s3-cache:local .     ##
##############################################################

FROM scratch

COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

COPY release/vela-s3-cache /bin/vela-s3-cache

ENTRYPOINT [ "/bin/vela-s3-cache" ]
