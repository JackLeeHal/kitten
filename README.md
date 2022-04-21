# kitten

[![Go](https://github.com/JackLeeHal/kitten/actions/workflows/go.yml/badge.svg)](https://github.com/JackLeeHal/kitten/actions/workflows/go.yml)

## What is kitten
Kitten 是一个为大规模小文件存储而生的分布式文件系统，主要架构参考了[Facebook的Haystack](https://www.usenix.org/legacy/event/osdi10/tech/full_papers/Beaver.pdf)，
并且从[bfs](https://github.com/Terry-Mao/bfs)中学习了很多优化手段。

## Features

## Quick Start

## Introduction
文件系统的优化方向一般都是从`怎么写`、`怎么存`、`怎么读`这三个方向入手。
1. 写：传统的机械硬盘由于有寻道和旋转这样的机械动作，顺序写入的性能是远大于随机写入的，所以Kitten的写入设计为顺序append。
2. 存&读：传统基于POSIX的文件系统中所有的文件都会有metadata，里面会有一些像permission，访问时间等数据，在大量小文件的情况下，这些元数据占用了过多的无用空间。
并且每次读取一个文件需要先做一次IO找到元数据，再通过元数据找到真正的文件。所以为了减少IO次数，Kitten将所有小文件append到一个大文件里，这里引入两个概念Superblock和Needle，
Superblock就是一个超大块，集合了顺序写入的小文件，Needle就是其中的每个小文件，读取时只需要通过内存里面维护的每个Needle的offset和size就能找到对应的文件。


## Roadmap
| Name                     | Issue                                               | Description                                                                     |
|--------------------------|-----------------------------------------------------|---------------------------------------------------------------------------------|
| Kitten's basic component | [#1](https://github.com/JackLeeHal/kitten/issues/1) | Implement basic component including `Store`, `Cache`, `Directory`               |
| Introduce Etcd           | [#2](https://github.com/JackLeeHal/kitten/issues/2) | Introduce Etcd for distributed management.                                      |
| Expose easy APIs         | [#3](https://github.com/JackLeeHal/kitten/issues/3) | Find an elegantly way to expose APIs.                                           |
| Support S3 API           | [#4](https://github.com/JackLeeHal/kitten/issues/4) | As S3 APIs are the de facto standards for OSS. Consider suppport S3 style APIs. |
| Implement erasure code   | [#4](https://github.com/JackLeeHal/kitten/issues/5) | Split data into two groups(hot/warm), use erasure code to store warm data.      |
