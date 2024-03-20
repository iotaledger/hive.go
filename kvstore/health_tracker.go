package kvstore

import "github.com/iotaledger/hive.go/ierrors"

const (
	// StoreVersionNone is used to load an existing store without a version check (e.g. in the tools).
	StoreVersionNone byte = 0
)

type StoreVersionUpdateFunc func(oldVersion byte, newVersion byte) error

var (
	ErrStoreVersionCheckNotSupported  = ierrors.New("store version check not supported")
	ErrStoreVersionUpdateFuncNotGiven = ierrors.New("store version update function not given")
)

type StoreHealthTracker struct {
	store                  KVStore
	storeVersion           byte
	storeVersionUpdateFunc StoreVersionUpdateFunc
}

func NewStoreHealthTracker(store KVStore, storePrefixHealth []byte, storeVersion byte, storeVersionUpdateFunc StoreVersionUpdateFunc) (*StoreHealthTracker, error) {
	healthStore, err := store.WithRealm(storePrefixHealth)
	if err != nil {
		return nil, err
	}

	s := &StoreHealthTracker{
		store:                  healthStore,
		storeVersion:           storeVersion,
		storeVersionUpdateFunc: storeVersionUpdateFunc,
	}

	if storeVersion != StoreVersionNone {
		if err := s.setStoreVersion(storeVersion); err != nil {
			return nil, err
		}
	}

	return s, nil
}

func (s *StoreHealthTracker) MarkCorrupted() error {
	if err := s.store.Set([]byte("dbCorrupted"), []byte{}); err != nil {
		return ierrors.New("failed to set store health status")
	}

	return s.store.Flush()
}

func (s *StoreHealthTracker) MarkTainted() error {
	if err := s.store.Set([]byte("dbTainted"), []byte{}); err != nil {
		return ierrors.New("failed to set store health status")
	}

	return s.store.Flush()
}

func (s *StoreHealthTracker) MarkHealthy() error {
	if err := s.store.Delete([]byte("dbCorrupted")); err != nil {
		return ierrors.New("failed to set store health status")
	}

	return nil
}

func (s *StoreHealthTracker) IsCorrupted() (bool, error) {
	contains, err := s.store.Has([]byte("dbCorrupted"))
	if err != nil {
		return true, ierrors.New("failed to read store health status")
	}

	return contains, nil
}

func (s *StoreHealthTracker) IsTainted() (bool, error) {
	contains, err := s.store.Has([]byte("dbTainted"))
	if err != nil {
		return true, ierrors.New("failed to read store health status")
	}

	return contains, nil
}

// StoreVersion returns the store version.
func (s *StoreHealthTracker) StoreVersion() (byte, error) {
	value, err := s.store.Get([]byte("dbVersion"))
	if err != nil {
		return 0, ierrors.New("failed to read store health version")
	}

	if len(value) < 1 {
		return 0, ierrors.New("failed to read store health version")
	}

	return value[0], nil
}

func (s *StoreHealthTracker) setStoreVersion(version byte) error {
	_, err := s.store.Get([]byte("dbVersion"))
	if ierrors.Is(err, ErrKeyNotFound) {
		// Only create the entry, if it doesn't exist already (fresh store)
		if err := s.store.Set([]byte("dbVersion"), []byte{version}); err != nil {
			return ierrors.New("failed to set store health version")
		}
	}

	return nil
}

func (s *StoreHealthTracker) CheckCorrectStoreVersion() (bool, error) {
	if s.storeVersion == StoreVersionNone {
		return false, ErrStoreVersionCheckNotSupported
	}

	storeVersion, err := s.StoreVersion()
	if err != nil {
		return false, err
	}

	return storeVersion == s.storeVersion, nil
}

// UpdateStoreVersion tries to migrate the existing data to the new store version.
// Returns true if the store needs to be updated / was updated.
func (s *StoreHealthTracker) UpdateStoreVersion() (bool, error) {
	storeVersion, err := s.StoreVersion()
	if err != nil {
		return false, err
	}

	if storeVersion == s.storeVersion {
		// already up to date
		return false, nil
	}

	if s.storeVersionUpdateFunc == nil {
		return true, ErrStoreVersionUpdateFuncNotGiven
	}

	if err := s.storeVersionUpdateFunc(storeVersion, s.storeVersion); err != nil {
		return true, err
	}

	if err := s.store.Set([]byte("dbVersion"), []byte{s.storeVersion}); err != nil {
		return true, ierrors.New("failed to set store health version")
	}

	return true, nil
}

func (s *StoreHealthTracker) Flush() error {
	return s.store.Flush()
}

func (s *StoreHealthTracker) Close() error {
	return s.store.Close()
}
