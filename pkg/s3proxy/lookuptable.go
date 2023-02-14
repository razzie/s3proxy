package s3proxy

import (
	"crypto/sha512"
	"encoding/binary"
	"math/rand"
)

var defaultTable = func() (table [256]byte) {
	for i := range table {
		table[i] = byte(i)
	}
	return
}()

type LookupTable struct {
	encrypt [256]byte
	decrypt [256]byte
}

func NewLookupTable(key string) *LookupTable {
	lt := &LookupTable{
		encrypt: defaultTable,
		decrypt: defaultTable,
	}
	if len(key) > 0 {
		rand.New(keyToSeed(key)).Shuffle(len(lt.decrypt), func(i, j int) {
			lt.decrypt[i], lt.decrypt[j] = lt.decrypt[j], lt.decrypt[i]
		})
		for i, b := range lt.decrypt {
			lt.encrypt[int(b)] = byte(i)
		}
	}
	return lt
}

func (lt *LookupTable) Encrypt(p []byte) {
	for i, b := range p {
		p[i] = lt.encrypt[b]
	}
}

func (lt *LookupTable) Decrypt(p []byte) {
	for i, b := range p {
		p[i] = lt.decrypt[b]
	}
}

func keyToSeed(key string) rand.Source {
	hash := sha512.Sum512([]byte(key))
	seed := binary.LittleEndian.Uint64(hash[:])
	return rand.NewSource(int64(seed))
}
