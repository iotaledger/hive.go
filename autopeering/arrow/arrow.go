package arrow

import (
	"crypto/md5"
	"encoding/binary"
	"github.com/iotaledger/hive.go/identity"
	"math/rand"
	"sync"
	"time"

	pb "github.com/iotaledger/hive.go/autopeering/arrow/proto"
	"google.golang.org/protobuf/proto"
)

const EPOCH_DURATION_SECONDS = 3600

// ArRow encapsulates high level functions around values management.
type ArRow struct {
	ars            []float64 // value of ars and rows
	rows           []float64 // value of ars and rows
	expirationTime time.Time // expiration time of the salt
	mutex          sync.RWMutex
}

// NewArRow generates a new values given a lifetime duration for given identity and number of neighbours
func NewArRow(lifetime time.Duration, k int, identity *identity.Identity) (arrowObj *ArRow, err error) {
	epochId := make([]byte, 8)
	now := time.Now().Unix()
	epoch := uint64(now - now%EPOCH_DURATION_SECONDS)
	binary.LittleEndian.PutUint64(epochId, epoch)

	h := md5.New()
	var seed = binary.BigEndian.Uint64(h.Sum(append(identity.ID().Bytes(), epochId...)))
	randSource := rand.New(rand.NewSource(int64(seed)))
	ars := make([]float64, 0, k)
	rows := make([]float64, 0, k)

	for idx := 0; idx < k; idx++ {
		ars = append(ars, randSource.Float64())
	}

	for idx := 0; idx < k; idx++ {
		rows = append(rows, randSource.Float64())
	}

	arrowObj = &ArRow{
		ars:  ars,
		rows: rows,

		expirationTime: time.Now().Add(lifetime),
	}
	return arrowObj, nil
}

func (s *ArRow) GetArs() []float64 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return append([]float64{}, s.ars...)
}
func (s *ArRow) GetRows() []float64 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return append([]float64{}, s.rows...)
}
func (s *ArRow) GetExpiration() time.Time {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.expirationTime
}

// Expired returns true if the given salt expired
func (s *ArRow) Expired() bool {
	return time.Now().After(s.GetExpiration())
}

// ToProto encodes the Salt into a proto buffer Salt message
func (s *ArRow) ToProto() *pb.ArRow {
	return &pb.ArRow{
		Ars:     s.ars,
		Rows:    s.rows,
		ExpTime: uint64(s.expirationTime.Unix()),
	}
}

// FromProto decodes a given proto buffer Salt message (in) and returns the corresponding Salt.
func FromProto(in *pb.ArRow) (*ArRow, error) {
	//if l := len(in.ArRow()); l != SaltByteSize {
	//	return nil, fmt.Errorf("invalid salt length: %d, need %d", l, SaltByteSize)
	//}
	out := &ArRow{
		ars:            in.GetArs(),
		rows:           in.GetRows(),
		expirationTime: time.Unix(int64(in.GetExpTime()), 0),
	}
	return out, nil
}

// Marshal serializes a given salt (s) into a slice of bytes (data)
func (s *ArRow) Marshal() ([]byte, error) {
	return proto.Marshal(s.ToProto())
}

// Unmarshal de-serializes a given slice of bytes (data) into a Salt.
func Unmarshal(data []byte) (*ArRow, error) {
	s := &pb.ArRow{}
	if err := proto.Unmarshal(data, s); err != nil {
		return nil, err
	}
	return FromProto(s)
}
