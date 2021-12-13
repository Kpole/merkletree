package MerkleTree

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"hash"
	"testing"
)

/*
定义基于 SHA256 以及 MD5 两种加密算法的 Content 类型
它们都有 CalculateHash 和 Equal 方法
 */

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

func (t TestSHA256Content) Equal(other Content) (bool, error) {
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

func (t TestMD5Content) Equal(other Content) (bool, error) {
	return t.x == other.(TestMD5Content).x, nil
}


func calHash(hash []byte, hashStrategy func() hash.Hash) ([]byte, error) {
	h := hashStrategy()
	if _, err := h.Write(hash); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}


var table = []struct {
	testCaseId			int
	hashStrategy		func() hash.Hash
	hashStrategyName	string
	contents			[]Content
	expectedHash		[]byte
	notInContents		Content
}{
	{
		testCaseId:          0,
		hashStrategy:        sha256.New,
		hashStrategyName:    "sha256",
		contents: []Content{
			TestSHA256Content{
				x: "Hello",
			},
			TestSHA256Content{
				x: "Hi",
			},
			TestSHA256Content{
				x: "Hey",
			},
			TestSHA256Content{
				x: "Hola",
			},
		},
		notInContents: TestSHA256Content{x: "NotInTestTable"},
		expectedHash:  []byte{95, 48, 204, 128, 19, 59, 147, 148, 21, 110, 36, 178, 51, 240, 196, 190, 50, 178, 78, 68, 187, 51, 129, 240, 44, 123, 165, 38, 25, 208, 254, 188},
	},
	{
		testCaseId:          1,
		hashStrategy:        md5.New,
		hashStrategyName:    "md5",
		contents: []Content{
			TestMD5Content{
				x: "123",
			},
			TestMD5Content{
				x: "234",
			},
			TestMD5Content{
				x: "345",
			},
			TestMD5Content{
				x: "456",
			},
			TestMD5Content{
				x: "1123",
			},
			TestMD5Content{
				x: "2234",
			},
			TestMD5Content{
				x: "3345",
			},
			TestMD5Content{
				x: "4456",
			},
			TestMD5Content{
				x: "5567",
			},
		},
		notInContents: TestMD5Content{x: "NotInTestTable"},
		expectedHash:  []byte{158, 85, 181, 191, 25, 250, 251, 71, 215, 22, 68, 68, 11, 198, 244, 148},
	},
}

// test：构建树，并对比根节点哈希值
func TestNewTree(t *testing.T) {
	for i := 0; i < len(table); i ++ {
		tree, err := newTree(table[i].contents, table[i].hashStrategy)
		if err != nil {
			t.Errorf("[case:%d] error: unexpected error: %v", table[i].testCaseId, err)
		}
		if bytes.Compare(tree.MerkleRootHash(), table[i].expectedHash) != 0 {
			t.Errorf("[case:%d] error: expected hash equal to %v got %v", table[i].testCaseId, table[i].expectedHash, tree.MerkleRootHash())
		}
	}
}

// test：调用verify方法检测是否通过，修改roothash后再次检测
func TestMerkleTree_VerifyTree(t *testing.T) {
	for i := 0; i < len(table); i++ {
		var tree *MerkleTree
		var err error
		tree, err = newTree(table[i].contents, table[i].hashStrategy)
		if err != nil {
			t.Errorf("[case:%d] error: unexpected error: %v", table[i].testCaseId, err)
		}
		v1, err := tree.verify()
		if err != nil {
			t.Fatal(err)
		}
		if v1 != true {
			t.Errorf("[case:%d] error: expected tree to be valid", table[i].testCaseId)
		}
		tree.Root.Hash = []byte{1}
		tree.RootHash = []byte{1}
		v2, err := tree.verify()
		if err != nil {
			t.Fatal(err)
		}
		if v2 != false {
			t.Errorf("[case:%d] error: expected tree to be invalid", table[i].testCaseId)
		}
	}
}

// test: 验证content，分三类：在树中（true）、在树中但修改了树根（false）、不在树中（false）
func TestMerkleTree_VerifyContent(t *testing.T) {
	for i := 0; i < len(table); i++ {
		var tree *MerkleTree
		var err error
		tree, err = newTree(table[i].contents, table[i].hashStrategy)
		if err != nil {
			t.Errorf("[case:%d] error: unexpected error: %v", table[i].testCaseId, err)
		}
		if len(table[i].contents) > 0 {
			v, err := tree.verifyContent(table[i].contents[0])
			if err != nil {
				t.Fatal(err)
			}
			if !v {
				t.Errorf("[case:%d] error: expected valid content", table[i].testCaseId)
			}
		}
		if len(table[i].contents) > 1 {
			v, err := tree.verifyContent(table[i].contents[1])
			if err != nil {
				t.Fatal(err)
			}
			if !v {
				t.Errorf("[case:%d] error: expected valid content", table[i].testCaseId)
			}
		}
		if len(table[i].contents) > 2 {
			v, err := tree.verifyContent(table[i].contents[2])
			if err != nil {
				t.Fatal(err)
			}
			if !v {
				t.Errorf("[case:%d] error: expected valid content", table[i].testCaseId)
			}
		}
		if len(table[i].contents) > 0 {
			tree.Root.Hash = []byte{1}
			tree.RootHash = []byte{1}
			v, err := tree.verifyContent(table[i].contents[0])
			if err != nil {
				t.Fatal(err)
			}
			if v {
				t.Errorf("[case:%d] error: expected invalid content", table[i].testCaseId)
			}
			if err := tree.rebuild(); err != nil {
				t.Fatal(err)
			}
		}
		v, err := tree.verifyContent(table[i].notInContents)
		if err != nil {
			t.Fatal(err)
		}
		if v {
			t.Errorf("[case:%d] error: expected invalid content", table[i].testCaseId)
		}
	}
}

// test: 对每个在树上的 content 进行Merkle证明
func TestMerkleTree_MerklePath(t *testing.T) {
	for i := 0; i < len(table); i++ {
		var tree *MerkleTree
		var err error
		tree, err = newTree(table[i].contents, table[i].hashStrategy)
		if err != nil {
			t.Errorf("[case:%d] error: unexpected error: %v", table[i].testCaseId, err)
		}
		for j := 0; j < len(table[i].contents); j++ {
			merklePath, index, _ := tree.getPath(table[i].contents[j])

			hash, err := tree.Leafs[j].calculateHash()
			if err != nil {
				t.Errorf("[case:%d] error: calculateNodeHash error: %v", table[i].testCaseId, err)
			}
			h := sha256.New()
			for k := 0; k < len(merklePath); k++ {
				if index[k] == 1 {
					hash = append(hash, merklePath[k]...)
				} else {
					hash = append(merklePath[k], hash...)
				}
				if _, err := h.Write(hash); err != nil {
					t.Errorf("[case:%d] error: Write error: %v", table[i].testCaseId, err)
				}
				hash, err = calHash(hash, table[i].hashStrategy)
				if err != nil {
					t.Errorf("[case:%d] error: calHash error: %v", table[i].testCaseId, err)
				}
			}
			if bytes.Compare(tree.MerkleRootHash(), hash) != 0 {
				t.Errorf("[case:%d] error: expected hash equal to %v got %v", table[i].testCaseId, hash, tree.MerkleRootHash())
			}
		}
	}
}
