package Trie

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
)

var (
	EmptyNodeRaw = []byte{}
	EmptyNodeHash, _ = hex.DecodeString("56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421")
)

type Proof interface {
	Put(key []byte, value Node) error
	Delete(key []byte) error
	Has(key []byte) (bool, error)
	Get(key []byte) (Node, error)
}

type DB struct {
	kv map[string]Node
}

func NewDB() *DB {
	return &DB{
		kv: make(map[string]Node),
	}
}

func keyS(key []byte) string { return fmt.Sprintf("%x", key) }

func (db *DB) Put(key []byte, value Node) error {
	db.kv[keyS(key)] = value
	return nil
}

func (db *DB) Delete(key []byte) error {
	delete(db.kv, keyS(key))
	return nil
}

func (db *DB) Has(key []byte) (bool, error) {
	_, ok := db.kv[keyS(key)]
	return ok, nil
}

func (db *DB) Get(key []byte) (Node, error) {
	val, ok := db.kv[keyS(key)]
	if !ok {
		return Node{}, errors.New("not found")
	}
	return val, nil
}



type Node struct {
	hash    []byte
	Branch 	[26][]byte
	Value	string
}

func (node *Node) Hash() ([]byte, error){
	h := sha256.New()
	for _, key := range node.Branch {
		if len(key) != 0{
			if _, err := h.Write(key); err != nil {
				return nil, err
			}
		}
	}
	if node.Value != "" {
		h.Write([]byte(node.Value))
	}
	node.hash = h.Sum(nil)
	return node.hash, nil
}

func (node *Node) verifyHash(db *DB) ([]byte, error) {
	h := sha256.New()
	for _, key := range node.Branch {
		if len(key) != 0 {
			flag, err := db.Has(key)
			if flag || err != nil {
				return nil, err
			}
			son, err := db.Get(key)
			if err != nil {
				return nil, err
			}
			sonBytes, err := son.verifyHash(db)
			if err != nil {
				return nil, err
			}
			if _, err := h.Write(sonBytes); err != nil {
				return nil, err
			}
		}
	}
	h.Write([]byte(node.Value))
	return h.Sum(nil), nil
}

func (node *Node) Update(key string, value string, db *DB) ([]byte, error) {
	if len(key) == 0 {
		node.Value = value
	} else {
		db.Delete(node.hash)
		var c int = (int)(key[0] - 'a')
		flag, err := db.Has(node.Branch[c])
		if err != nil {
			return nil, err
		}
		if flag == false {
			son := &Node{}
			node.Branch[c], err = son.Update(key[1:], value, db)
			if err != nil {
				return nil, err
			}
		} else {
			son, err := db.Get(node.Branch[c])
			if err != nil {
				return nil, err
			}
			node.Branch[c], err = son.Update(key[1:], value, db)
			if err != nil {
				return nil, err
			}
		}
	}
	hash, err := node.Hash()
	if err != nil {
		return nil, err
	}
	db.Put(hash, *node)
	return hash, nil
}

type Trie struct {
	root Node
}

func NewTrie() *Trie {
	return &Trie {}
}

// 返回根 hash
func (t *Trie) Hash() []byte {
	return t.root.hash
}

func (t *Trie) Get(key string, db *DB) (string, bool, error) {
	node := t.root
	for i := 0; i < len(key); i ++ {
		c := key[i] - 'a'
		if flag, err := db.Has(node.Branch[c]); !flag || err != nil {
			return "", false, err
		}
		son, err := db.Get(node.Branch[c])
		if err != nil {
			return "", false, err
		}
		if i == len(key) - 1 {
			if son.Value == "" {
				return "", false, nil
			} else {
				return son.Value, true, nil
			}
		}
		node = son
	}
	return "", false, nil
}

func (t *Trie) Put(key string, value string, db *DB) error {
	if _, err := t.root.Update(key, value, db); err != nil {
		return err
	}
	return nil
}


func (t *Trie) verifyTrie(db *DB) (bool, error) {
	calcRootHash, err := t.root.verifyHash(db)
	if err != nil {
		return false, nil
	}
	if bytes.Compare(t.root.hash, calcRootHash) == 0 {
		return true, nil
	}
	return false, nil
}

func (t *Trie) proof(key string, db *DB)(*DB, bool) {
	proofdb := NewDB()
	node := t.root
	if len(key) == 0 {
		return nil, false
	}
	for {
		proofdb.Put(node.hash, node)
		if len(key) == 0 {
			return proofdb, len(node.Value) != 0
		}
		c := key[0] - 'a'
		key = key[1:]
		if len(node.Branch[c]) == 0 {
			return nil, false
		}
		flag, err := db.Has(node.Branch[c])
		if err != nil || flag == false {
			return nil, false
		}
		node, err = db.Get(node.Branch[c])
		if err != nil {
			return nil, false
		}
	}
}

func verifyProof(rootHash []byte, key string, proofdb *DB) (value string, err error) {
	targetHash := rootHash
	for i := 0; ; i++ {
		if flag, err := proofdb.Has(targetHash); err != nil || flag == false {
			return "", fmt.Errorf("proof node %d (hash %064x) missing", i, targetHash)
		}
		node, err := proofdb.Get(targetHash)
		if err != nil {
			return "", fmt.Errorf("proof node %d (hash %064x) missing", i, targetHash)
		}
		if i == len(key) {
			return node.Value, nil
		}
		c := key[i] - 'a'
		targetHash = node.Branch[c]
	}
}












