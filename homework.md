目标：实现Merkle Tree + 实现一种优化。

优化种类：

- 交易无序
- 无法查询（结合字典树）
- 快速过滤（结合布隆过滤器）
- 快速过滤（结合累加器）

提交内容：
1. 代码
  - Merkle Tree 代码
  - 一种优化

2. 论文
  - Merkle Tree 介绍
  - 优化 Merkle Tree 出发点
  - 优化方案
  - 优化效果
  - 总结

Merkle Tree 是用作快速归纳和效验大规模数据完整性的树形数据结构。

1. Introduced by Ralph Merkle, 1979
  - "Classic" cryptographic construction
  - Involves combining hash functions on binary tree structure
2. A public-key authentication scheme
  - Using only one-way hash function as building blocks
  - Also public-key signatures (Lamport’s one-time signatures)
3. Construct binary tree over data values
  - The root and internal nodes are the hash value of its two children.
  - Authenticate a sequence of data values $D_0,D_1,\cdots,D_n$

Beneficial feature: any change on the data will be reflected at the tree root

- 提供了一种证明数据完整性/有效性的手段。
- 可以利用很少的内存/磁盘空间实现高效的计算
- 它的论证与管理只需要少量的网络传输流量


### Bloom Filter

https://developer.aliyun.com/article/773205

