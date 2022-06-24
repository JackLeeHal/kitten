# kitten🐱

[![Go](https://github.com/JackLeeHal/kitten/actions/workflows/go.yml/badge.svg)](https://github.com/JackLeeHal/kitten/actions/workflows/go.yml)

## What is kitten
Kitten 是一个为大规模小文件存储而生的分布式文件系统，核心架构参考了[Facebook的Haystack](https://www.usenix.org/legacy/event/osdi10/tech/full_papers/Beaver.pdf)。 （本项目只是一个学习项目，未经过生产环境验证）

## Features

* 高吞吐、低延时的百万级文件处理
* 容错性，在不同地区镜像文件
* 高性价比
* 架构简单

## Quick Start

更多的设计和实现细节可在 [docs](https://github.com/JackLeeHal/kitten/docs) 文件夹找到.

## Roadmap
| Name                     | Issue                                               | Description                                                                |
|--------------------------|-----------------------------------------------------|----------------------------------------------------------------------------|
| Kitten's basic component | [#1](https://github.com/JackLeeHal/kitten/issues/1) | Implement basic component including `Store`, `Cache`, `Directory`          |
| Introduce Etcd           | [#2](https://github.com/JackLeeHal/kitten/issues/2) | Introduce Etcd for distributed management.                                 |
| Expose easy APIs         | [#3](https://github.com/JackLeeHal/kitten/issues/3) | Find an elegantly way to expose APIs.                                      |
| Support S3 API           | [#4](https://github.com/JackLeeHal/kitten/issues/4) | As S3 APIs are the de facto standards for OSS， support S3 style APIs.      |
| Implement erasure code   | [#5](https://github.com/JackLeeHal/kitten/issues/5) | Split data into two groups(hot/warm), use erasure code to store warm data. |

## 感谢

kitten 受这些项目启发:

[bfs](https://github.com/Terry-Mao/bfs) distributed file system(small file storage) writen in golang.

[seaweedfs](https://github.com/chrislusf/seaweedfs) SeaweedFS is a fast distributed storage system for blobs, objects, files, and data lake, for billions of files!

再次感谢 Jetbrains 的开源 Licence 赞助!

![](/docs/jb_square.svg) 