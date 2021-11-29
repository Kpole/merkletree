package src

import (
	"bytes"
	"errors"
	"fmt"
	"hash"
)

type Content interface {
	CalculateHash() ([] byte, error)
	Equal(other Content) (bool, error)
}

type Node struct {
	Tree		*MerkleTree
	Parent 		*Node
	LeftSon		*Node
	RightSon 	*Node
	isLeaf		bool
	isDup 		bool
	Hash   		[]byte
	C			Content
}

type MerkleTree struct {
	Root			*Node
	RootHash 		[]byte
	Leafs			[]*Node
	hashStrategy 	func() hash.Hash
}

func newTree(cs []Content, hashStrategy func() hash.Hash) (*MerkleTree, error) {
	t := &MerkleTree {
		hashStrategy: hashStrategy,
	}
	root, leafs, err := build(cs, t)
	if err != nil {
		return nil, err
	}
	t.Root = root
	t.Leafs = leafs
	t.RootHash = root.Hash
	return t, nil
}

func build(cs []Content, t *MerkleTree)(*Node, []*Node, error) {
	if len(cs) == 0 {
		return nil, nil, errors.New("error: cannot build tree without content")
	}
	var leafs []*Node
	for _,c := range cs {
		hash, err := c.CalculateHash()
		if err != nil {
			return nil, nil, err
		}
		leafs = append(leafs, &Node{
			Hash: 	hash,
			C:			c,
			isLeaf:	true,
			Tree:		t,
		})
	}
	if len(leafs) % 2 == 1 {
		duplicate := &Node {
			Hash: 	leafs[len(leafs)-1].Hash,
			C:			leafs[len(leafs)-1].C,
			isLeaf: true,
			isDup : true,
			Tree: 	t,
		}
		leafs = append(leafs, duplicate)
	}
	root, err := buildIntermediate(leafs, t)
	if err != nil {
		return nil, nil, err
	}
	return root, leafs, nil
}

func buildIntermediate(nodeList []*Node, t *MerkleTree) (*Node, error) {
	var nodes []*Node
	for i := 0; i < len(nodeList); i += 2 {
		hash := t.hashStrategy()
		var left, right int = i, i + 1
		if i + 1 == len(nodeList) {
			right = i
		}
		contentHash := append(nodeList[left].Hash, nodeList[right].Hash...)
		if _,err := hash.Write(contentHash); err != nil {
			return nil, err
		}
		newNode := &Node {
			LeftSon: 	nodeList[left],
			RightSon: nodeList[right],
			Hash:			hash.Sum(nil),
			Tree:			t,
		}
		nodes = append(nodes, newNode)
		nodeList[left].Parent = newNode
		nodeList[right].Parent = newNode
		if len(nodeList) == 2 {
			return newNode, nil
		}
	}
	return buildIntermediate(nodes, t)
}

/*
Node 类型方法
String() 打印基本属性
verifyNodeHash() 重新递归计算节点哈希值
calculateHash() 重新计算节点的哈希值
*/
func (n *Node) String() string {
	return fmt.Sprintf("%t %t %v %s", n.isLeaf, n.isDup, n.Hash, n.C)
}

func (n *Node) verifyHash() ([] byte, error) {
	if n.isLeaf {
		return n.C.CalculateHash()
	}
	rightBytes, err := n.RightSon.verifyHash()
	if err != nil {
		return nil, err
	}
	leftBytes, err := n.LeftSon.verifyHash()
	if err != nil {
		return nil, err
	}
	h := n.Tree.hashStrategy()
	if _, err := h.Write(append(leftBytes, rightBytes...)); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

func (n *Node) calculateHash() ([]byte, error) {
	if n.isLeaf {
		return n.C.CalculateHash()
	}
	h := n.Tree.hashStrategy()
	if _,err := h.Write(append(n.LeftSon.Hash, n.RightSon.Hash...)); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

/*
MerkleTree 类型方法
String() : 打印所有叶子节点信息
MerkleRootHash() 返回树根哈希值
getPath(content) 如果 content 存在，返回路径 
verify() 验证根哈希
*/

func (m *MerkleTree) String() string {
	s := ""
	for _, l := range m.Leafs {
		s += fmt.Sprint(l);
		s += "\n"
	}
	return s
}

func (m *MerkleTree) MerkleRootHash() []byte {
	return m.RootHash
}

func (m *MerkleTree) getPath(content Content)([][]byte, []int64, error) {
	for _, current := range m.Leafs {
		ok, err := current.C.Equal(content)
		if err != nil {
			return nil, nil, err
		}
		if ok {
			currentParent := current.Parent
			var path [][]byte
			var index []int64
			for currentParent != nil {
				if bytes.Equal(currentParent.LeftSon.Hash, current.Hash) {
					path = append(path, currentParent.RightSon.Hash)
					index = append(index, 1)
				} else {
					path = append(path, currentParent.LeftSon.Hash)
					index = append(index, 0)
				}
				current = current.Parent
				currentParent = current.Parent
			}
			return path, index, nil
		}
	}
	return nil, nil, nil
}

func (m *MerkleTree) verify() (bool, error) {
	calculatedMerkleRoot, err := m.Root.verifyHash()
	if err != nil {
		return false, err
	}
	if bytes.Compare(m.RootHash, calculatedMerkleRoot) == 0 {
		return true, nil
	}
	return false, nil
}

func (m *MerkleTree) verifyContent(content Content) (bool, error) {
	h := m.hashStrategy()
	for _, l := range(m.Leafs) {
		ok, err := l.C.Equal(content)
		if err != nil {
			return false, err
		}
		if ok {
			currentParent := l.Parent
			for currentParent != nil {
				rightBytes, err := currentParent.RightSon.calculateHash()
				if err != nil {
					return false, nil
				}
				leftBytes, err := currentParent.LeftSon.calculateHash()
				if err != nil {
					return false, nil
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

func (m *MerkleTree) rebuildWithContent(cs []Content) error {
	root, leafs, err := build(cs, m) 
	if err != nil {
		return err
	}
	m.Root = root
	m.Leafs = leafs
	m.RootHash = root.Hash
	return nil
}


func (m *MerkleTree) rebuild() error {
	var cs []Content
	for _, c := range m.Leafs {
		cs = append(cs, c.C)
	}
	root, leafs, err := build(cs, m)
	if err != nil {
		return err
	}
	m.Root = root
	m.Leafs = leafs
	m.RootHash = root.Hash
	return nil
}