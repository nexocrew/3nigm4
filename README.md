3n4
======
_“if something happens to me” let the computer act on my behalf..._

3n4 is an open source project aimed at offering a secure storage solution to sharing information with selected recipients even if  the original data owner were to lose the files. 3n4 is mainly developed in [Go](https://golang.org/).

<a href="https://asciinema.org/a/07pxxtloh42kdygx7jz8sbuen" target="_blank"><img src="https://asciinema.org/a/07pxxtloh42kdygx7jz8sbuen.png" height="550" width="700"/></a>

## Status
[![Build Status](https://travis-ci.org/nexocrew/3nigm4.svg?branch=develop)](https://travis-ci.org/nexocrew/3nigm4)
[![GoDoc](https://godoc.org/github.com/nexocrew/3nigm4?status.svg)](https://godoc.org/github.com/nexocrew/3nigm4)
[![GitHub issues](https://img.shields.io/github/issues/nexocrew/3nigm4.svg "GitHub issues")](https://github.com/nexocrew/3nigm4)
[![Dev chat at https://gitter.im/nexocrew/3nigm4-framework](https://img.shields.io/badge/gitter-dev_chat-46bc99.svg)](https://gitter.im/nexocrew/3nigm4-framework?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

Table of contents
---------------------

 - [Context](#context)
 - [Description](#description)
 - [Components](#components)
 - [Installation](#installation)
 - [Testing](#testing)
 - [Documentation](#documentation)
 - [Contribution](#contribution)
 - [Contributors](#contributors)

## Context
In the last few years we have unfortunately seen many different examples of researchers, journalists or political opponents being arrested or killed while operating in totalitarian regimes, often also having their research work destroyed or shelved. These kind of occurrences cannot be predicted with certainty and can happen very suddenly. Starting from these facts, we started thinking about how to provide a solution to those people, giving them the opportunity to securely share critical informations with third parties, even while in prison or deceased: this is the main objective of the 3n4 project.

## Description
3n4 (or 3nigm4) is an open source project providing secure data storing and unattended sharing.

Behind the scenes we have designed 3n4 to use PGP keys and AES encryption and to manage actual data and all critical metadata in totally separate files, passing through different processing flows. Furthermore the architecture is intended to be distributed, segmented (organized in micro-services) and lightweight.

While the security community's aim is more often focused on vulnerability detection, this project is focused on the solution: we are driven by the idea of providing a tool usable by anyone (not only by sec researchers or crypto fanatics) but  featuring a security level able to satisfy both usage scenarios.

## Components
The 3n4 project is composed of two main components:
- **A storage service (3n4store)**: storing data after removing all metadata from uploaded files (they are actually chunked and anonymized). Based on the use of PGP and AES and maintaining, client side, all critical infos about the uploaded file;
- **A “If Something Happens To Me" service (3n4ishtm)**: the service will let any user upload sensible data references (encrypted using PGP using the recipient’s keys and referring to files using the previously described storage service) and to deliver it, to pre-defined recipients, if the legitimate data owner do not ping the system after a specific amount of time;
- **A command line client tool (3n4cli)**: the client to access all 3n4 services.

## Installation
All 3n4 services are available as docker images (official Docker Hub):

- **3n4auth**: docker pull nexo/3n4auth
- **3n4store**: docker pull nexo/3n4store
- **3n4cli**: docker pull nexo/3n4cli

Each image have a specific README file reporting specific deployment instructions and requirements.

## Testing
We are right now in alpha with the first version of all componets. To have access to the test instance and try using the 3n4 client please contact [@dyst0ni3](https://github.com/dystonie): you'll be provided with test credentials.

## Docs
In the `docs/` directory you can find useful documentation.

## Contribution

The project is in the early development stages: contributors are welcome! Please before contributing read the [issues](https://github.com/nexocrew/3nigm4/issues).
Thanks!

## Contributors
[@dyst0ni3](https://github.com/dystonie)
[@FredMaggiowski](https://github.com/federicomaggi)
[@Bestbug](https://github.com/bestbug456)
