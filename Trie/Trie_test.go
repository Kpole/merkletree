package Trie

import (
	"bytes"
	"testing"
)

/*
Test1：测试Get、Put操作
 */
func TestGetPut(t *testing.T) {
	// 尝试访问不存在的 key，得到 false
	t.Run("should get nothing if key does not exist", func(t *testing.T) {
		trie := NewTrie()
		db := NewDB()
		_, found, err := trie.Get("notexist", db)
		if err != nil {
			t.Errorf("Error:%v", err)
		}
		if found {
			t.Errorf("Error")
		}
	})

	// 尝试访问存在的键值，希望成功获得期望的结果
	t.Run("should get value if key exist", func(t *testing.T) {
		trie := NewTrie()
		db := NewDB()
		trie.Put("hi", "hello", db)
		val, found, err := trie.Get("hi", db)
		if err != nil {
			t.Errorf("Error: %v", err)
		}
		if found == false || val != "hello" {
			t.Errorf("should get value hello")
		}
	})

	// 将Trie中的key-value更新之后，希望获得更新过的值
	t.Run("should get updated value", func(t *testing.T) {
		trie := NewTrie()
		db := NewDB()
		trie.Put("world", "hello", db)
		trie.Put("world", "world", db)
		val, found, err := trie.Get("world", db)
		if err != nil {
			t.Errorf("%v", err)
		}
		if found != true || val != "world" {
			t.Errorf("shoud get updated value")
		}
	})
}

/*
Test2：检测数据是否正确，即包含同样key-value集合的Trie树的根哈希值是否相等
 */
func TestDataIntegrity(t *testing.T) {
	// 构造三个不同的Trie树，希望其树根哈希互不相同
	t.Run("should get a different hash if a new key-value pair was added or updated", func(t *testing.T) {
		db := NewDB()
		trie := NewTrie()
		hash0 := trie.Hash()

		trie.Put("abcd", "hello", db)
		hash1 := trie.Hash()

		trie.Put("ab", "world", db)
		hash2 := trie.Hash()

		trie.Put("ab", "test", db)
		hash3 := trie.Hash()
		//fmt.Printf("hash0: %v\n", hash0)
		//fmt.Printf("hash1: %v\n", hash1)
		//fmt.Printf("hash2: %v\n", hash2)
		//fmt.Printf("hash3: %v\n", hash3)
		//fmt.Printf("compare hash2 with hash3: %v\n", bytes.Equal(hash2, hash3))
		if bytes.Equal(hash0, hash1) || bytes.Equal(hash1, hash2) || bytes.Equal(hash2, hash3) {
			t.Errorf("should different")
		}
	})


	// 构建两个Trie树，以不同的顺序插入相同的集合，希望根哈希一致
	t.Run("should get the same hash if two tries have the identicial key-value pairs", func(t *testing.T) {
		trie1 := NewTrie()
		db1 := NewDB()
		trie1.Put("ab", "world", db1)
		trie1.Put("abcd", "hello", db1)

		trie2 := NewTrie()
		db2 := NewDB()
		trie2.Put("abcd", "hello", db2)
		trie2.Put("ab", "world", db2)

		hash1 := trie1.Hash()
		hash2 := trie2.Hash()

		if bytes.Equal(hash1,hash2) != true {
			t.Errorf("should equial")
		}
	})
}

/*
Test3: 测试Merkle证明
 */
func TestProveAndVerifyProof(t *testing.T) {
	// 尝试在Trie树中查找不存在的key，希望证明返回false
	t.Run("should not generate proof for non-exist key", func(t *testing.T) {
		tr := NewTrie()
		trdb := NewDB()
		tr.Put("abc", "hello", trdb)
		tr.Put("abcde", "world", trdb)
		notExistKey := "abcd"
		_, ok := tr.proof(notExistKey,trdb)
		if ok == true {
			t.Errorf("should not generate proof for non-exist key")
		}
	})

	// 查询存在的数据，并希望通过根哈希可以通过该路径得到数据
	t.Run("should generate a proof for an existing key, the proof can be verified with the merkle root hash", func(t *testing.T) {
		tr := NewTrie()
		trdb := NewDB()
		tr.Put("abc", "hello", trdb)
		tr.Put("abcde", "world", trdb)

		key := "abcde"
		proofdb, ok := tr.proof(key, trdb)
		if ok != true {
			t.Errorf("abc should ok")
		}

		rootHash := tr.Hash()

		// verify the proof with the root hash, the key in question and its proof
		val, err := verifyProof(rootHash, key, proofdb)
		if err != nil {
			t.Errorf("err should no err")
		}
		if val != "world" {
			t.Errorf("val should be hello")
		}
	})

	// 查询Trie中存在的数据，使用错误的树根哈希不能通过Merkle路径得到该数据。
	t.Run("should fail the verification if the trie was updated", func(t *testing.T) {
		tr := NewTrie()
		trdb := NewDB()
		tr.Put("abc", "hello", trdb)
		tr.Put("abcde", "world", trdb)

		// Trie 更新之前的树根哈希
		rootHash := tr.Hash()

		// 更新Trie，然后尝试证明 "abc"
		tr.Put("efg", "trie", trdb)
		key := "abc"
		proofdb, ok := tr.proof(key, trdb)
		if ok == false {
			t.Errorf("abc should ok")
		}

		// 根哈希不匹配，验证失败
		_, err := verifyProof(rootHash, key, proofdb)
		if err == nil {
			t.Errorf("err should err")
		}
	})
}

