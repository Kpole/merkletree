package Trie

import (
	"encoding/hex"
	"errors"
	"fmt"
	"crypto/sha256"
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
		if _, err := h.Write(key); err != nil {
			return nil, err
		}
	}
	node.hash = h.Sum(nil)
	return node.hash, nil
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

func (t *Trie) Put(key string, value string) {

}












