package flake

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/sigurn/crc8"
)

const (
	v0 = iota << 60
	v1
	tsMask = 0x0fffffffffffff00
	vMask  = 0xf000000000000000
)

var crcT = crc8.MakeTable(crc8.CRC8)

// Flake produces unique time sortable IDs.
//
// Flake v1 are 12 byte long and are comprised of:
//  0  3    7       15       23       31
// +----+----+--------+--------+--------+
// |0001|                               |
// +----+  Unix Time in usec   +--------+
// |                           |Overflow|
// +---------+--------+--------+--------+
// |        Worker ID          |  CRC8  |
// +---------+--------+--------+--------+
//
type Flake interface {
	Next() []byte
	NextHex() string
	NextB64() string
}

type flake struct {
	lock     sync.Mutex
	buf, enc []byte
	counter  uint64
	lts      uint64
}

// New creates new Flake v1 instance from a given worker ID.
// To avoid collisions in case of multiple Flakes started with the same `wid`
// overflow counter is started from a strongly random uint8.
//
// It is caller responsibility to ensure that worker IDs are 24 bit long.
func New(wid uint32) Flake {
	buf := make([]byte, 12)
	buf[8] = uint8(wid >> 16)
	buf[9] = uint8(wid >> 8)
	buf[10] = uint8(wid)

	// Start counter with crypto random value
	tmp := make([]byte, 1, 1)
	_, err := rand.Read(tmp)
	if err != nil {
		panic(err)
	}

	return &flake{
		buf:     buf,
		enc:     make([]byte, hex.EncodedLen(len(buf))),
		counter: uint64(tmp[0]),
	}
}

// Next produces new unique ID.
func (f *flake) Next() []byte {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.tick()
	ret := make([]byte, len(f.buf))
	copy(ret, f.buf)
	return ret
}

// NextHex produces new unique ID hex encoded into string.
func (f *flake) NextHex() string {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.tick()
	hex.Encode(f.enc, f.buf)
	return string(f.enc)
}

// NextB64 produces new unique ID base64 encoded into string.
func (f *flake) NextB64() string {
	f.lock.Lock()
	defer f.lock.Unlock()

	f.tick()
	base64.URLEncoding.Encode(f.enc[:16], f.buf)

	return string(f.enc[:16])
}

func (f *flake) tick() {
	ts := uint64(time.Now().UTC().UnixNano())
	ts = ts / 1000
	ts = ts << 8
	ts = v1 | ts | f.counter
	if ts <= f.lts {
		f.counter = (f.counter + 1) & 0xff
	}
	// fmt.Printf("%064b\n", v1|ts|f.counter)
	f.lts = ts
	binary.BigEndian.PutUint64(f.buf[:8], v1|ts|f.counter)
	f.buf[11] = crc8.Checksum(f.buf[:11], crcT)
}

// Validate checks given slice to conform to Flake structure.
func Validate(id []byte) error {
	ts := binary.BigEndian.Uint64(id[:8])

	// if tt := time.Unix(0, int64(ts&tsMask>>8*1000)); time.Since(tt) > time.Millisecond {
	// 	return fmt.Errorf("time %v is more than 1ms off of now %v", tt, time.Now())
	// }

	if v := ts >> 60; v != 1 {
		return fmt.Errorf("expected Flake v1, got v%v", v)
	}

	if lid := len(id); lid != 12 {
		return fmt.Errorf("expected length 12, got %v", lid)
	}

	if fcs := crc8.Checksum(id[:11], crcT); fcs != id[11] {
		return fmt.Errorf("CRC8 mismatch %x != %x", fcs, id[11])

	}
	return nil
}
