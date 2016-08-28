3nigm4
======
_A command line chat based on keybase._


3nigm4 is a [Go](https://golang.org/) application developed by _nexocrew_. The purpose of the software is to provide a GPG-based framework, integrated with [keybase](https://keybase.io) as a trusted key server, offering secure chat and file sharing capabilities.

## Status
[![Build Status](https://travis-ci.org/nexocrew/3nigm4.svg?branch=develop)](https://travis-ci.org/nexocrew/3nigm4)
[![GoDoc](https://godoc.org/github.com/nexocrew/3nigm4?status.svg)](https://godoc.org/github.com/nexocrew/3nigm4)
[![GitHub issues](https://img.shields.io/github/issues/nexocrew/3nigm4.svg "GitHub issues")](https://github.com/nexocrew/3nigm4)
[![Dev chat at https://gitter.im/nexocrew/3nigm4-framework](https://img.shields.io/badge/gitter-dev_chat-46bc99.svg)](https://gitter.im/nexocrew/3nigm4-framework?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

## Test coverage
[![Lib crypto coverage](http://gocover.io/_badge/github.com/nexocrew/3nigm4/lib/crypto?0 "lib crypto coverage")](http://gocover.io/github.com/nexocrew/3nigm4/lib/crypto) /
[![Lib filemanager coverage](http://gocover.io/_badge/github.com/nexocrew/3nigm4/lib/filemanager?0 "lib filemanager coverage")](http://gocover.io/github.com/nexocrew/3nigm4/lib/filemanager) /
[![Lib workingqueue coverage](http://gocover.io/_badge/github.com/nexocrew/3nigm4/lib/workingqueue?0 "lib workingqueue coverage")](http://gocover.io/github.com/nexocrew/3nigm4/lib/workingqueue) /
[![Lib storageclient coverage](http://gocover.io/_badge/github.com/nexocrew/3nigm4/lib/storageclient?0 "lib storageclient coverage")](http://gocover.io/github.com/nexocrew/3nigm4/lib/storageclient) /
[![Lib s3 coverage](http://gocover.io/_badge/github.com/nexocrew/3nigm4/lib/s3?0 "lib storageclient coverage")](http://gocover.io/github.com/nexocrew/3nigm4/lib/s3)

Table of contents
---------------------

 - [Components](#components)
 - [Installation](#installation)
 - [Testing](#testing)
 - [Benchmark](#benchmark)
 - [Documentation](#documentation)
 - [Contribution](#contribution)
 - [Contributors](#contributors)

## Components
The software is designed to be deployed as a microservices architecture. The components designed for the _first alpha_ are: 

- **3n4chat**: The service exposes REST APIs to exchange chat information. It will store the conversation encrypted and unaccessible to the server itself. (More information in the _docs_)
- **3n4store** This service will expose REST APIs to implement a authenticated interface to an S3 backend storage. All passed data will be encrypted client side, separated in chuncks of fixed size and separated from the encryption keys.
- **3n4auth** This service will provide authentication capabilities for the previously presented backend services.
- **3n4cli**: The clinet side command line interface.

## Installation
All 3nigm4 components are available as docker images (official Docker Hub):

- **3n4auth**: docker pull nexo/3n4auth
- **3n4store**: docker pull nexo/3n4store
- **3n4cli**: docker pull nexo/3n4cli

Each image have a specific README file reporting specific deployment instructions and requirements.

## Testing
_Todo_


## Benchmark
_Todo_


## Docs
In the `docs/` directory you can find useful documentation.

## Contribution

All contributions are well received. Please before contributing read the [issues](https://github.com/nexocrew/3nigm4/issues), we'll use them to keep track of todos, fixes and everything that the project will need.
Thanks!

## Contributors
[@dyst0ni3](https://github.com/dystonie)
[@FredMaggiowski](https://github.com/federicomaggi)
[@Bestbug](https://github.com/bestbug456)
