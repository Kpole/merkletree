package Trie

import (
	"bytes"
	"testing"
)

func TestGetPut(t *testing.T) {
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

func TestDataIntegrity(t *testing.T) {
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

	t.Run("should get the same hash if two tries have the identicial key-value pairs", func(t *testing.T) {
		trie1 := NewTrie()
		db1 := NewDB()
		trie1.Put("abcd", "hello", db1)
		trie1.Put("ab", "world", db1)

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


func TestProveAndVerifyProof(t *testing.T) {
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

	t.Run("should fail the verification if the trie was updated", func(t *testing.T) {
		tr := NewTrie()
		trdb := NewDB()
		tr.Put("abc", "hello", trdb)
		tr.Put("abcde", "world", trdb)

		// the hash was taken before the trie was updated
		rootHash := tr.Hash()

		// the proof was generated after the trie was updated
		tr.Put("efg", "trie", trdb)
		key := "abc"
		proofdb, ok := tr.proof(key, trdb)
		if ok == false {
			t.Errorf("abc should ok")
		}

		// should fail the verification since the merkle root hash doesn't match
		_, err := verifyProof(rootHash, key, proofdb)
		if err == nil {
			t.Errorf("err should err")
		}
	})
}

