package src

import (
	"crypto/md5"
	"crypto/sha256"
)

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

