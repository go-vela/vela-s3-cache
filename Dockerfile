
# SPDX-License-Identifier: Apache-2.0

#############################################################################
##     docker build --no-cache --target certs -t vela-s3-cache:certs .     ##
#############################################################################

FROM alpine:3.23.3@sha256:25109184c71bdad752c8312a8623239686a9a2071e8825f20acb8f2198c3f659 AS certs

RUN apk add --update --no-cache ca-certificates

##############################################################
##     docker build --no-cache -t vela-s3-cache:local .     ##
##############################################################

FROM scratch

COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

COPY release/vela-s3-cache /bin/vela-s3-cache

ENTRYPOINT [ "/bin/vela-s3-cache" ]
