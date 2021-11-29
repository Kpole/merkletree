

### MerkleTree 代码解读

### 数据结构

#### Content

包含 Content 接口的类型可以被当做 Tree 中的一个节点。可以计算哈希、判断值是否相等。

```go
type Content interface {
	CalculateHash() ([]byte, error)
	Equals(other Content) (bool, error)
}
```

#### MerkleTree

Merkle Tree 类型，包括指向根的指针、根哈希值、叶子节点指针数组、以及所采用的 Hash 函数。

```go
type MerkleTree struct {
	Root         *Node
	merkleRoot   []byte
	Leafs        []*Node
	hashStrategy func() hash.Hash
}
```

#### Node

Node 类型可以代表树的根、树内部的节点或者是叶子节点。它内部包含了指向树的指针（MerkleTree*类型）、指向双亲、左右孩子的指针、表示是否是叶子节点的标记变量、哈希值、Content值（如果是一个叶子的话）、其他元数据。

```go
type Node struct {
	Tree   *MerkleTree
	Parent *Node
	Left   *Node
	Right  *Node
	leaf   bool
	dup    bool
	Hash   []byte
	C      Content
}
```

### 方法

#### 全局方法

##### NewTree

此处一次介绍三个函数

1. `NewTree(cs []Content) (*MerkleTree, error)` 传入 Content 类型的数组 cs，以此构建 Tree。函数里面将哈希方法设置为 `sha256.New`，然后将cs数组递归构建树的任务交给了 `buildWithContent`
2. `buildWithContent(cs []Content, t *MerkleTree) (*Node, []*Node, error)`，将 cs 的内容依次计算 Hash，放入新创建的 Node 对象，然后 append 到局部变量 leafs 数组中。如果 leafs 的数组长度是奇数，那么需要复制最后一个节点使得 leafs 长度变为偶数。最后将 leafs 交给 `buildIntermediate` 函数去构建树里面的其他结点
3. `buildIntermediate(nl []*Node, t *MerkleTree) (*Node, error)` 自底向上的构建树，参数 nl 数组中，每相邻的两个点构成上一层节点的左右孩子。然后递归的向树深度更小的结点构建。

只需要传入 cs数组，就可以返回一个指向构建好的 MerkleTree 的指针。在调用第二个以及第三个函数的过程中，指向树的指针 `t` 需要传入其中，除了获取提前定义好的哈希函数，另外一个目的是使得树中的所有结点都可以使用指针 Tree 直接指向树。

```go
func NewTree(cs []Content) (*MerkleTree, error) {
	var defaultHashStrategy = sha256.New
	t := &MerkleTree{
		hashStrategy: defaultHashStrategy,
	}
	root, leafs, err := buildWithContent(cs, t)
	if err != nil {
		return nil, err
	}
	t.Root = root
	t.Leafs = leafs
	t.merkleRoot = root.Hash
	return t, nil
}
func buildWithContent(cs []Content, t *MerkleTree) (*Node, []*Node, error) {
	if len(cs) == 0 {
		return nil, nil, errors.New("error: cannot construct tree with no content")
	}
	var leafs []*Node
	for _, c := range cs {
		hash, err := c.CalculateHash()
		if err != nil {
			return nil, nil, err
		}
		leafs = append(leafs, &Node{
			Hash: hash,
			C:    c,
			leaf: true,
			Tree: t,
		})
	}
	if len(leafs)%2 == 1 {
		duplicate := &Node{
			Hash: leafs[len(leafs)-1].Hash,
			C:    leafs[len(leafs)-1].C,
			leaf: true,
			dup:  true,
			Tree: t,
		}
		leafs = append(leafs, duplicate)
	}
	root, err := buildIntermediate(leafs, t)
	if err != nil {
		return nil, nil, err
	}

	return root, leafs, nil
}
func buildIntermediate(nl []*Node, t *MerkleTree) (*Node, error) {
	var nodes []*Node
	for i := 0; i < len(nl); i += 2 {
		h := t.hashStrategy()
		var left, right int = i, i + 1
		if i+1 == len(nl) {
			right = i
		}
		chash := append(nl[left].Hash, nl[right].Hash...)
		if _, err := h.Write(chash); err != nil {
			return nil, err
		}
		n := &Node{
			Left:  nl[left],
			Right: nl[right],
			Hash:  h.Sum(nil),
			Tree:  t,
		}
		nodes = append(nodes, n)
		nl[left].Parent = n
		nl[right].Parent = n
		if len(nl) == 2 {
			return n, nil
		}
	}
	return buildIntermediate(nodes, t)
}
```

如果需要传参指定哈希函数的种类，可以使用：

```go
func NewTreeWithHashStrategy(cs []Content, hashStrategy func() hash.Hash) (*MerkleTree, error) {
	t := &MerkleTree{
		hashStrategy: hashStrategy,
	}
	root, leafs, err := buildWithContent(cs, t)
	if err != nil {
		return nil, err
	}
	t.Root = root
	t.Leafs = leafs
	t.merkleRoot = root.Hash
	return t, nil
}
```



#### Node 类型的方法

##### String

返回节点的内容（格式化之后）

```go
func (n *Node) String() string {
	return fmt.Sprintf("%t %t %v %s", n.leaf, n.dup, n.Hash, n.C)
}
```



##### verifyNode

计算 Node 所表示节点的哈希值。其过程是递归的计算该节点子树中所有的点的哈希。

```go
func (n *Node) verifyNode() ([]byte, error) {
	if n.leaf {
		return n.C.CalculateHash()
	}
	rightBytes, err := n.Right.verifyNode()
	if err != nil {
		return nil, err
	}

	leftBytes, err := n.Left.verifyNode()
	if err != nil {
		return nil, err
	}

	h := n.Tree.hashStrategy()
	if _, err := h.Write(append(leftBytes, rightBytes...)); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}
```

##### calculateNodeHash

计算 Node 所表示节点的哈希。其方法是直接链接其所指左右孩子的哈希并计算。

```go
func (n *Node) calculateNodeHash() ([]byte, error) {
	if n.leaf {
		return n.C.CalculateHash()
	}
	h := n.Tree.hashStrategy()
	if _, err := h.Write(append(n.Left.Hash, n.Right.Hash...)); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}
```



#### MerkleTree 类型的方法

##### String

返回MerkleTree中所有叶子节点的内容

```go
func (m *MerkleTree) String() string {
	s := ""
	for _, l := range m.Leafs {
		s += fmt.Sprint(l)
		s += "\n"
	}
	return s
}
```

##### MerkleRoot

返回树根哈希

```go
func (m *MerkleTree) MerkleRoot() []byte {
	return m.merkleRoot
}
```

##### GetMerklePath

传入 Content，寻找 MerkleTree 中到达存放该内容的叶子节点的路径。返回值有两个，第一个从叶子到根的路径上每个节点的哈希值，另一个是路径上每个节点的索引（左右孩子）。注意，这里保存的并非路径上经过的节点，而是相应的兄弟节点，目的是方便校验根的哈希是否正确。

```go
func (m *MerkleTree) GetMerklePath(content Content) ([][]byte, []int64, error) {
	for _, current := range m.Leafs {
		ok, err := current.C.Equals(content)
		if err != nil {
			return nil, nil, err
		}

		if ok {
			currentParent := current.Parent
			var merklePath [][]byte
			var index []int64
			for currentParent != nil {
                // 如果与父亲的左孩子哈希一致，那么会append右孩子的哈希值，并且将 index 追加 1
				if bytes.Equal(currentParent.Left.Hash, current.Hash) {
					merklePath = append(merklePath, currentParent.Right.Hash)
					index = append(index, 1) // right leaf
				} else {
					merklePath = append(merklePath, currentParent.Left.Hash)
					index = append(index, 0) // left leaf
				}
				current = currentParent
				currentParent = currentParent.Parent
			}
			return merklePath, index, nil
		}
	}
	return nil, nil, nil
}
```

##### VerifyTree

递归的去重新计算树根哈希，然后与树根存储的哈希值对比，相等则返回 true，否则返回 false

```go
func (m *MerkleTree) VerifyTree() (bool, error) {
	calculatedMerkleRoot, err := m.Root.verifyNode()
	if err != nil {
		return false, err
	}

	if bytes.Compare(m.merkleRoot, calculatedMerkleRoot) == 0 {
		return true, nil
	}
	return false, nil
}
```

##### VerifyContent

检查 Content 是否在 Tree 中，如果存在，则进一步检查其叶子节点到Root的路径上哈希值是否正确。

```go
func (m *MerkleTree) VerifyContent(content Content) (bool, error) {
	for _, l := range m.Leafs {
		ok, err := l.C.Equals(content)
		if err != nil {
			return false, err
		}

		if ok {
			currentParent := l.Parent
			for currentParent != nil {
				h := m.hashStrategy()
				rightBytes, err := currentParent.Right.calculateNodeHash()
				if err != nil {
					return false, err
				}

				leftBytes, err := currentParent.Left.calculateNodeHash()
				if err != nil {
					return false, err
				}

				if _, err := h.Write(append(leftBytes, rightBytes...)); err != nil {
					return false, err
				}
				if bytes.Compare(h.Sum(nil), currentParent.Hash) != 0 {
					return false, nil
				}
				currentParent = currentParent.Parent
			}
			return true, nil
		}
	}
	return false, nil
}
```

##### RebuildTree

重新构建树，将树上叶子节点中的 content 加入 cs 数组，然后调用 `buildWithContent` 函数

```go
func (m *MerkleTree) RebuildTree() error {
	var cs []Content
	for _, c := range m.Leafs {
		cs = append(cs, c.C)
	}
	root, leafs, err := buildWithContent(cs, m)
	if err != nil {
		return err
	}
	m.Root = root
	m.Leafs = leafs
	m.merkleRoot = root.Hash
	return nil
}
```

##### RebuildTreeWith

传入新的 Content 去构建 MerkleTree

```go
func (m *MerkleTree) RebuildTreeWith(cs []Content) error {
	root, leafs, err := buildWithContent(cs, m)
	if err != nil {
		return err
	}
	m.Root = root
	m.Leafs = leafs
	m.merkleRoot = root.Hash
	return nil
}
```



### 测试

测试中使用两种哈希函数，分别是 `SHA256` 和 `MD5`。需要为他们构建相应的类型，并实现 `CalculateHash`和`Equals` 方法。

```go
type TestSHA256Content struct {
	x string
}

func (t TestSHA256Content) CalculateHash() ([]byte, error) {
	h := sha256.New()
	if _, err := h.Write([]byte(t.x)); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

func (t TestSHA256Content) Equals(other Content) (bool, error) {
	return t.x == other.(TestSHA256Content).x, nil
}

type TestMD5Content struct {
	x string
}

func (t TestMD5Content) CalculateHash() ([]byte, error) {
	h := md5.New()
	if _, err := h.Write([]byte(t.x)); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

func (t TestMD5Content) Equals(other Content) (bool, error) {
	return t.x == other.(TestMD5Content).x, nil
}
```



计算hash函数，需要传入字节数组、哈希策略（即指定的哈希函数）

```go
func calHash(hash []byte, hashStrategy func() hash.Hash) ([]byte, error) {
	h := hashStrategy()
	if _, err := h.Write(hash); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}
```

测试过程中，提前准备好了多组数组放入到 table 中。以下只介绍其中一个例子，其他例子思路类似。

#### TestNewTree

取出 table 中的每个使用默认哈希策略的数据，构建MerkleTree 后对比根结点哈希与测试数据期望得到的哈希。

```go
func TestNewTree(t *testing.T) {
	for i := 0; i < len(table); i++ {
		if !table[i].defaultHashStrategy {
			continue
		}
		tree, err := NewTree(table[i].contents)
		if err != nil {
			t.Errorf("[case:%d] error: unexpected error: %v", table[i].testCaseId, err)
		}
		if bytes.Compare(tree.MerkleRoot(), table[i].expectedHash) != 0 {
			t.Errorf("[case:%d] error: expected hash equal to %v got %v", table[i].testCaseId, table[i].expectedHash, tree.MerkleRoot())
		}
	}
}
```

