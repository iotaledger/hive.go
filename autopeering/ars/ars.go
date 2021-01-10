package ars

import (
	"crypto/md5"
	"encoding/binary"
	"github.com/iotaledger/hive.go/identity"
	"math/rand"
	"sync"
	"time"

	pb "github.com/iotaledger/hive.go/autopeering/ars/proto"
	"google.golang.org/protobuf/proto"
)

const EPOCH_DURATION_SECONDS = 3600

// Ars encapsulates high level functions around ars management.
type Ars struct {
	ars            []float64 // value of ars
	expirationTime time.Time // expiration time of the salt
	mutex          sync.RWMutex
}

// NewArs generates a new ars given a lifetime duration for given identity and number of neighbours
func NewArs(lifetime time.Duration, k int, identity *identity.Identity) (arsObj *Ars, err error) {
	epochId := make([]byte, 8)
	now := time.Now().Unix()
	epoch := uint64(now - now%EPOCH_DURATION_SECONDS)
	binary.LittleEndian.PutUint64(epochId, epoch)

	h := md5.New()
	var seed = binary.BigEndian.Uint64(h.Sum(append(identity.ID().Bytes(), epochId...)))
	randSource := rand.New(rand.NewSource(int64(seed)))
	ars := make([]float64, 0, k)

	for idx := 0; idx < k; idx++ {
		ars = append(ars, randSource.Float64())
	}

	arsObj = &Ars{
		ars:            ars,
		expirationTime: time.Now().Add(lifetime),
	}
	return arsObj, nil
}

func (s *Ars) GetArs() []float64 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return append([]float64{}, s.ars...)
}

func (s *Ars) GetExpiration() time.Time {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.expirationTime
}

// Expired returns true if the given salt expired
func (s *Ars) Expired() bool {
	return time.Now().After(s.GetExpiration())
}

// ToProto encodes the Salt into a proto buffer Salt message
func (s *Ars) ToProto() *pb.Ars {
	return &pb.Ars{
		Ars:     s.ars,
		ExpTime: uint64(s.expirationTime.Unix()),
	}
}

// FromProto decodes a given proto buffer Salt message (in) and returns the corresponding Salt.
func FromProto(in *pb.Ars) (*Ars, error) {
	//if l := len(in.Ars()); l != SaltByteSize {
	//	return nil, fmt.Errorf("invalid salt length: %d, need %d", l, SaltByteSize)
	//}
	out := &Ars{
		ars:            in.GetArs(),
		expirationTime: time.Unix(int64(in.GetExpTime()), 0),
	}
	return out, nil
}

// Marshal serializes a given salt (s) into a slice of bytes (data)
func (s *Ars) Marshal() ([]byte, error) {
	return proto.Marshal(s.ToProto())
}

// Unmarshal de-serializes a given slice of bytes (data) into a Salt.
func Unmarshal(data []byte) (*Ars, error) {
	s := &pb.Ars{}
	if err := proto.Unmarshal(data, s); err != nil {
		return nil, err
	}
	return FromProto(s)
}
