3nigm4
======

_A command line chat based on keybase._


3nigm4 is a [Go](https://golang.org/) application developed by _nexocrew_. The purpose of the software is to provide users a GPG-based chat using [keybase](https://keybase.io) as a trusted key server.

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

- **3nigm4-server**: The server exposes REST APIs to exchange chat information. It will store the conversation encrypted and unaccessible to the server itself. (More information in the _docs_)
- **3nigm4-cli**: The command line human-machine interface.
- **da3nigm4**: A daemon which communicates with the _client_, the _server_ and the keybase server. It also processes core, storage and cryptographic operations.

## Installation
_Todo_

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
