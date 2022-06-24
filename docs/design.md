# Kitten Design Documentation

## Introduction
When the traditional file system stores a large number of small files, there will be an IO bottleneck of metadata, because each time you read a file,
you need to do IO first to find the metadata, and then find the real file through the metadata.And the data such as permission and access time stored in metadata may be useless.

In the case of a large number of small files, the metadata size corresponding to the data you save may be similar to your data size, resulting in a lot of space waste.

Kitten optimized this phenomenon in two directions：
1. Sequential writing: because the traditional mechanical hard disk has mechanical actions such as seek and rotation, the performance of sequential writing is much greater than that of random writing, so kitten's writing is designed as sequential append.
2. Metadata：Kitten appends all small files to a large file to reduce metadata，Two concepts superblock and needle are introduced here,
   Superblock is a super block that collects small files written in sequence. Need is each small file in it. When reading, you only need to find the corresponding file through the offset and size of each need maintained in memory.

Kitten is suitable for files that：`written once`, `read often`, `never modified`, and `rarely deleted`.
Goal of Kitten：`High throughput + low delay`, `Fault-tolerant`, `Cost-effective`, `Simple`.

Kitten includes the following modules：
![](docs/kitten.png)

### Store

### Proxy

As a user oriented module, the proxy module shields various operations inside kitten and exposes three simple APIs, `get`, `post` and `delete`. Represent read, write and delete operations respectively. The proxy communicates downward through grpc.

### Directory

### Cache