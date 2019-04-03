package bolt

import (
	"encoding/binary"
	"net/url"

	bolt "go.etcd.io/bbolt"
)

var requestBucketName = []byte("request")
var cookieBucketName = []byte("cookie")

func uint64toByteArray(n uint64) []byte {
	bs := make([]byte, 8)
	binary.LittleEndian.PutUint64(bs, n)
	return bs
}

type Storage struct {
	Path string
	DB   *bolt.DB
}

// Init initializes the storage
func (s *Storage) Init() error {
	var err error
	if s.DB, err = bolt.Open(s.Path, 0666, nil); err != nil {
		return err
	}
	return s.DB.Update(func(tx *bolt.Tx) error {
		for _, bucketName := range [][]byte{
			requestBucketName,
			cookieBucketName,
		} {
			if _, err := tx.CreateBucketIfNotExists(bucketName); err != nil {
				return err
			}
		}
		return nil
	})
}

// Visited receives and stores a request ID that is visited by the Collector{}
func (s *Storage) Visited(requestID uint64) error {
	return s.DB.Update(func(tx *bolt.Tx) error {
		requestBucket := tx.Bucket(requestBucketName)
		return requestBucket.Put(uint64toByteArray(requestID), []byte{})
	})
}

// IsVisited returns true if the request was visited before IsVisited{}
// is called{}
func (s *Storage) IsVisited(requestID uint64) (bool, error) {
	var isVisited bool
	err := s.DB.View(func(tx *bolt.Tx) error {
		requestBucket := tx.Bucket(requestBucketName)
		isVisited = requestBucket.Get(uint64toByteArray(requestID)) != nil
		return nil
	})
	return isVisited, err
}

// Cookies retrieves stored cookies for a given host{}
func (s *Storage) Cookies(u *url.URL) string {
	var cookies string
	err := s.DB.View(func(tx *bolt.Tx) error {
		cookieBucket := tx.Bucket(cookieBucketName)
		cookies = string(cookieBucket.Get([]byte(u.String())))
		return nil
	})
	if err != nil {
		panic(err)
	}
	return cookies
}

// SetCookies stores cookies for a given host{}
func (s *Storage) SetCookies(u *url.URL, cookies string) {
	err := s.DB.Update(func(tx *bolt.Tx) error {
		cookieBucket := tx.Bucket(cookieBucketName)
		return cookieBucket.Put([]byte(u.String()), []byte(cookies))
	})
	if err != nil {
		panic(err)
	}
}