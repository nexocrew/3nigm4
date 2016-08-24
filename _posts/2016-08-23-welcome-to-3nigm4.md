---
layout:     post
title:      Welcome to 3nigm4
date:       2016-08-23 15:31:19
author:     Danilo Bestbug
summary:    3nigm4: a command line chat based on keybase.
categories: jekyll
thumbnail:  heart
tags:
 - 3nigm4
 - summary
 - first contact
 - docs
---

3nigm4: a command line chat based on keybase.

3nigm4 is a [Go][1] application developed by nexocrew. The purpose of the software is to provide a GPG-based framework, integrated with [keybase][2] as a trusted key server, offering secure chat and file sharing capabilities.

The software is designed to be deployed as a microservices architecture. The components designed for the first alpha are:

    *3nigm4-chat-backend: The service exposes REST APIs to exchange chat information. It will store the conversation encrypted and unaccessible to the server itself. (More information in the docs)
    *3nigm4-storage-backend This service will expose REST APIs to implement a authenticated interface to an S3 backend storage. All passed data will be encrypted client side, separated in chuncks of fixed size and separated from the encryption keys.
    *3nigm4-auth-server This service will provide authentication capabilities for the previously presented backend services.
    *3nigm4-cli: The clinet side command line interface.
    *3nigm4-deamon: A daemon which communicates with the various services and will be controlled by the 3nigm4-cli application. It also processes core, storage and cryptographic operations.


[1]: https://golang.org/
[2]: https://keybase.io/
