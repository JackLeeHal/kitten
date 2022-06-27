
# Kitten Design Documentation


## Introduction
传统文件系统在存储大量小文件的情况下，会出现元数据的IO瓶颈，因为每次读取一个文件需要先做一次IO找到元数据，再通过元数据找到真正的文件。并且元数据中存储的像permission、访问时间等数据可能是无用的。

在小文件数量很大的情况下你存一个数据对应的元数据大小可能跟你的数据大小差不多，这样就造成了大量的空间浪费。

Kitten 从两个方向优化了这个现象：
1. 顺序写：传统的机械硬盘由于有寻道和旋转这样的机械动作，顺序写入的性能是远大于随机写入的，所以 Kitten 的写入设计为顺序 append。
2. 元数据方面：Kitten 将所有小文件 append 到一个大文件里，这里引入两个概念 Superblock 和 Needle，
   Superblock 就是一个超大块，集合了顺序写入的小文件，Needle 就是其中的每个小文件，读取时只需要通过内存里面维护的每个 Needle 的 offset 和 size 就能找到对应的文件。

Kitten 适合的文件特点是：`一次写入`，`从不更新`，`不定期会读`，`极少删除`.

Kitten 的设计目标是：`高吞吐+低延时`，`有容错机制`，`低成本`，`架构简单`.

围绕这些目标，Kitten 包含了以下几个模块：
![](kitten.png)

### Store

Store 作为整个系统最核心的模块，一个 `store` 由多个 `Superblock`组成，一个 `Superblock` 由多个 `Needle` 组成。

![](img.png)

![](img_1.png)

为了更快的读到所需的数据，每个 Store 还会维护一个内存中的数据结构，Key 是 Needle 的 Key，value 则是这个 Needle 的 offset 和 size。这样只需要持有一个 block 的 fd 就能找到里面的所有数据。

Store模块的pb表述如下：

```
service Store {
  rpc GetFile(GetFileRequest) returns (GetFileResponse) {}
  rpc UploadFile(UploadFileRequest) returns (UploadFileResponse) {}
  rpc DeleteFile(DeleteFileRequest) returns (DeleteFileResponse) {}
}
```

下面详细描述三种操作

#### 文件读取

Store 读取文件时接收到的参数有三个：vid， key， cookie（防止猜出文件地址 URL 的拼接方式的暴力攻击）。



vid 代表 Volume 的 id，Store 首先通过 vid 找到对应的 Superblock。然后再通过 key 找到对应 Needle 的 offset 和 size。这样只需要一次 ReadAt 就能拿到数据。

#### 文件写入

文件写入的时候只是比读取时多了一个文件真正的数据这个参数。Store 会写入一个 Superblock 然后更新内存中的map（注意写入是 append-only 的）。在做更新操作的时候，由于是顺序写入，所以需要更新内存中的map保证老的Needle不会被读取到。

#### 文件删除

文件删除只需要将 Needle 中的 Flag 设置为删除即可。请求已经删除的文件会返回一个错误。已删除的文件会暂时占用储存空间，后面会详细介绍怎么处理这种孤儿Needle。



### Proxy

Proxy 模块作为一个面向用户的模块，屏蔽了 Kitten 内部的各种操作，向外暴露三个简单的API，`get`、`post`和`delete`。分别代表读取、写入和删除操作。Proxy向下都是通过grpc进行通信。

### Directory

### Cache
