#!/bin/bash

#
# Example execution:
# sh docker_build.sh /home/user/programming/gopath/ github.com/nexocrew/3nigm4/ linux amd64
#

#
# Base image name
#
GOLANG_DEFAULT_IMAGE=nexo/golang:latest

#
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
# $5 golang image
#
function clean()	 {
	docker run --rm -v ${1}/src/:/usr/gopath/src/:rw \
	-w /usr/gopath/src/${2} -e GOPATH=/usr/gopath \
	-e GOARCH=${4} -e GOOS=${3} ${5} \
	bash -c "make clean"
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
# $5 golang image
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

	# -e GOBIN=/usr/src/${2}/${BUILD_PATH} 
	# install it!
	docker run --rm -v ${1}/src/:/usr/gopath/src/:rw \
	-v ${1}/src/${2}/${BUILD_PATH}:/usr/gopath/bin/:rw \
	-w /usr/gopath/src/${2} -e GOPATH=/usr/gopath \
	-e GOARCH=${4} -e GOOS=${3} \
	--name golangbuild ${5} \
	bash -c "make install"
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
		if [ ! -d ${GO_PATH} ] 
		then
			echo "Invalid path to the gopath directory, aborting."
			exit 1
		fi
		#
		# Get project relative path
		#
		PROJ_PATH=${2}
		if [ ! -d ${GO_PATH}/src/${PROJ_PATH} ] 
		then
			echo "Invalid path to the project directory, aborting."
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

		GOLANG=${GOLANG_DEFAULT_IMAGE}
		if [ ! -z ${ALT_GOLANG_IMAGE} ]
		then
			echo "Choose a different golang image: ${ALT_GOLANG_IMAGE}"
			GOLANG=${ALT_GOLANG_IMAGE}
		fi

		#
		# Execute clean
		#
		clean ${GO_PATH} ${PROJ_PATH} ${PLATFORM} ${ARCH} ${GOLANG}
		if [ $? -ne 0 ]
		then
			echo "Unable to clean the app."
			exit 1
		fi

		#
		# Execute build
		#
		build ${GO_PATH} ${PROJ_PATH} ${PLATFORM} ${ARCH} ${GOLANG}
		if [ $? -ne 0 ]
		then
			echo "Unable to install the app."
			exit 1
		fi

		#
		# Create tar archive
		#
		pushd ${GO_PATH}/src/${PROJ_PATH}/build/
		tar -czvf ${GO_PATH}/src/${PROJ_PATH}/build/3nigm4-${PLATFORM}-${ARCH}.tar ${PLATFORM}-${ARCH}/
		if [ $? -ne 0 ]
		then
			echo "Unable to tar the build dir."
			exit 1
		fi
		popd


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
