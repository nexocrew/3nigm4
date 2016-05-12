#!/bin/bash

GOLANG_IMAGE=golang:latest

#
# General-purpose environment variables:
#
# 	GCCGO
# 		The gccgo command to run for 'go build -compiler=gccgo'.
# 	GOARCH
# 		The architecture, or processor, for which to compile code.
# 		Examples are amd64, 386, arm, ppc64.
# 	GOBIN
# 		The directory where 'go install' will install a command.
# 	GOOS
# 		The operating system for which to compile code.
# 		Examples are linux, darwin, windows, netbsd.
# 	GOPATH
# 		See 'go help gopath'.
# 	GORACE
# 		Options for the race detector.
# 		See https://golang.org/doc/articles/race_detector.html.
# 	GOROOT
# 		The root of the go tree.
#

#
# clean <gopath> <src_dir> <target> <arch>
# $1 gopath
# $2 project path
# $3 target os
# $4 architecture type
#
function clean()	 {
	docker run --rm -v ${1}/src/:/usr/src/:Z \
	-w /usr/src/${2} -e GOPATH=/usr/:${GOPATH} \
	-e GOARCH=${4} -e GOOS=${3} ${GOLANG_IMAGE} \
	bash -c make clean
	if [ $? -ne 0 ]
	then
		return 1
	fi
	return 0
}

#
# build <gopath> <src_dir> <target> <arch>
# $1 gopath
# $2 project path
# $3 target os
# $4 architecture type
#
function build() {
	BUILD_PATH=build/${3}-${4}
	# check for build dir
	if [ -d ${1}/src/${2}/${BUILD_PATH} ]
	then
		mkdir -p ${1}src/${2}/${BUILD_PATH}
	else
		rm -Rf ${1}/src/${2}/${BUILD_PATH}/*
	fi

	# install it!
	docker run --rm -v ${1}/src/:/usr/src/:Z \
	-w /usr/src/${2} -e GOPATH=/usr/:${GOPATH} \
	-e GOBIN=/usr/src/${2}/${BUILD_PATH} -e GOARCH=${4} \
	-e GOOS=${3} ${GOLANG_IMAGE} \
	bash -c make install
	if [ $? -ne 0 ]
	then
		return 1
	fi
	return 0
}

#
# Verify docker presence on the system
#
if command -v docker >/dev/null 2>&1; 
then

	#
	# Check for arguments
	#
	if [ $# -eq 4 ] 
	then
		#
		# Get gopath
		#
		GO_PATH=${1}
		if [ -z ${GO_PATH} ] 
		then
			echo "Invalid path to the gopath directory, aborting."
			exit 1
		fi
		#
		# Get project relative path
		#
		PROJ_PATH=${2}
		if [ -z ${PROJ_PATH} ] 
		then
			echo "Invalid path to the project relative directory, aborting."
			exit 1
		fi
		#
		# Get target platform
		#
		PLATFORM=${3}
		if [ -z ${PLATFORM} ] 
		then
			echo "Invalid platform name, aborting."
			exit 1
		fi
		#
		# Get architecture
		#
		ARCH=${4}
		if [ -z ${ARCH} ] 
		then
			echo "Invalid architecture, aborting."
			exit 1
		fi

		#
		# Execute clean
		#
		clean ${GO_PATH} ${PROJ_PATH} ${PLATFORM} ${ARCH}
		if [ $? -ne 0 ]
		then
			echo "Unable to clean the app."
			exit 1
		fi

		#
		# Execute build
		#
		build ${GO_PATH} ${PROJ_PATH} ${PLATFORM} ${ARCH}
		if [ $? -ne 0 ]
		then
			echo "Unable to install the app."
			exit 1
		fi

		exit 0

	else
		echo "Unexpected number of arguments, having $# expecting 3."
		echo "sh docker_build.sh <gopath> <src_path> <os> <arch>"
		echo "Available os: linux, darwin, windows, netbsd."
		echo "Available arch: amd64, 386, arm, ppc64."
		exit 1
	fi

else
	echo "Docker is required to execute this script."
	exit 1
fi