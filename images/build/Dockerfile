#
# This file build a Docker image for build environment
# Target: 3nigm4 build and deploy
#
FROM golang:1.6-alpine
MAINTAINER Guido Ronchetti <guido.ronchetti@nexo.cloud>
ENV REFRESHED_AT 2016-07-27
LABEL nexo.product="3nigm4"

#
# Install all required tools 
#
RUN apk update && apk upgrade
RUN apk add --no-cache make bash