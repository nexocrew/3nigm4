#
# This file build a Docker image for production environment
# Target: 3nigm4 run images
#
FROM alpine:3.4
MAINTAINER Guido Ronchetti <guido.ronchetti@nexo.cloud>
ENV REFRESHED_AT 2016-07-27
LABEL nexo.product="3nigm4"

#
# Install all required tools 
#
RUN apk update && apk upgrade
RUN apk add curl ca-certificates bash && update-ca-certificates