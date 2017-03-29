package flake_test

import (
	"encoding/base64"
	"encoding/hex"
	"testing"

	"github.com/sigurn/crc8"

	flake "github.com/illotum/flake"
)

const AAx24 = 0x00aaaaaa

var crcT = crc8.MakeTable(crc8.CRC8)

func TestBytes(t *testing.T) {
	f := flake.New(AAx24)
	id := f.Next()
	err := flake.Validate(id)
	if err != nil {
		t.Error(err)
	}
}

func TestHex(t *testing.T) {
	f := flake.New(AAx24)
	idS := f.NextHex()
	id := make([]byte, 12)
	_, err := hex.Decode(id, []byte(idS))
	if err != nil {
		t.Error(err)
	}
	err = flake.Validate(id)
	if err != nil {
		t.Error(err)
	}
}

func TestB64(t *testing.T) {
	f := flake.New(AAx24)
	idS := f.NextB64()
	id := make([]byte, 12)
	_, err := base64.URLEncoding.Decode(id, []byte(idS))
	if err != nil {
		t.Error(err)
	}
	err = flake.Validate(id)
	if err != nil {
		t.Error(err)
	}
}

func BenchmarkBytes(b *testing.B) {
	f := flake.New(AAx24)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			unoptimize(f.Next())
		}
	})
}

func BenchmarkHex(b *testing.B) {
	f := flake.New(AAx24)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			unoptimize(f.NextHex())
		}
	})
}

func BenchmarkB64(b *testing.B) {
	f := flake.New(AAx24)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			unoptimize(f.NextB64())
		}
	})
}

func unoptimize(i interface{}) {}
