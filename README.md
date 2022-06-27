# kitten🐱

[![Go](https://github.com/JackLeeHal/kitten/actions/workflows/go.yml/badge.svg)](https://github.com/JackLeeHal/kitten/actions/workflows/go.yml)


[简体中文](README_zh.md)

## About kitten
Kitten is a distributed file system optimized for small files storage，core concepts based on [Facebook‘s Haystack](https://www.usenix.org/legacy/event/osdi10/tech/full_papers/Beaver.pdf)。 

## Features

* High throughput & Low latency for millions of files
* Fault-tolerant, store replicates files on different racks
* Cost-effective
* Simple

## Quick Start

Find more detail design in [docs](https://github.com/JackLeeHal/kitten/docs) folder.

## Roadmap
| Name                     | Issue                                               | Description                                                                |
|--------------------------|-----------------------------------------------------|----------------------------------------------------------------------------|
| Kitten's basic component | [#1](https://github.com/JackLeeHal/kitten/issues/1) | Implement basic component including ~~`Store`~~, `Cache`, `Directory`(WIP) |
| Introduce Etcd           | [#2](https://github.com/JackLeeHal/kitten/issues/2) | Introduce Etcd for distributed management(WIP).                            |
| Expose easy APIs         | [#3](https://github.com/JackLeeHal/kitten/issues/3) | Find an elegantly way to expose APIs.                                      |
| Support S3 API           | [#4](https://github.com/JackLeeHal/kitten/issues/4) | As S3 APIs are the de facto standards for OSS， support S3 style APIs.      |
| Implement erasure code   | [#5](https://github.com/JackLeeHal/kitten/issues/5) | Split data into two groups(hot/warm), use erasure code to store warm data. |

## Acknowledgments

Inspired by:

[bfs](https://github.com/Terry-Mao/bfs) distributed file system(small file storage) writen in golang.

[seaweedfs](https://github.com/chrislusf/seaweedfs) SeaweedFS is a fast distributed storage system for blobs, objects, files, and data lake, for billions of files!

Thanks again for JetBrains' Sponsorship!

![](/docs/jb_square.svg) 