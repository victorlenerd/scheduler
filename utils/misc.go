package utils

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/segmentio/ksuid"
	"github.com/spf13/afero"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"scheduler0/config"
	"unsafe"
)

func MakeDirIfNotExist(logger *log.Logger, path string) (string, bool) {
	dir, err := os.Getwd()
	if err != nil {
		logger.Fatalln(fmt.Errorf("Fatal error getting working dir: %s \n", err))
	}

	dirPath := fmt.Sprintf("%v/%v", dir, path)
	fs := afero.NewOsFs()

	exists, err := afero.DirExists(fs, dirPath)
	if err != nil {
		return dirPath, exists
	}

	if !exists {
		err := fs.Mkdir(dirPath, os.ModePerm)
		if err != nil {
			logger.Println("err", err)
			return dirPath, exists
		}
	}

	return dirPath, exists
}

func GetRandomSha256() string {
	randomId := ksuid.New().String()
	hash := sha256.New()
	hash.Write([]byte(randomId))
	return hex.EncodeToString(hash.Sum(nil))
}

func ReadUint64(b []byte) (uint64, error) {
	var sz uint64
	if err := binary.Read(bytes.NewReader(b), binary.LittleEndian, &sz); err != nil {
		return 0, err
	}
	return sz, nil
}

func WriteUint64(w io.Writer, v uint64) error {
	return binary.Write(w, binary.LittleEndian, v)
}

func BytesFromSnapshot(rc io.ReadCloser) ([]byte, error) {
	var uint64Size uint64
	inc := int64(unsafe.Sizeof(uint64Size))

	// Read all the data into RAM, since we have to decode known-length
	// chunks of various forms.
	var offset int64
	b, err := ioutil.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("readall: %s", err)
	}

	// Get size of database, checking for compression.
	compressed := false
	sz, err := ReadUint64(b[offset : offset+inc])
	if err != nil {
		return nil, fmt.Errorf("read compression check: %s", err)
	}
	offset = offset + inc

	if sz == math.MaxUint64 {
		compressed = true
		// Database is actually compressed, read actual size next.
		sz, err = ReadUint64(b[offset : offset+inc])
		if err != nil {
			return nil, fmt.Errorf("read compressed size: %s", err)
		}
		offset = offset + inc
	}

	// Now read in the database file data, decompress if necessary, and restore.
	var database []byte
	if sz > 0 {
		if compressed {
			buf := new(bytes.Buffer)
			gz, err := gzip.NewReader(bytes.NewReader(b[offset : offset+int64(sz)]))
			if err != nil {
				return nil, err
			}

			if _, err := io.Copy(buf, gz); err != nil {
				return nil, fmt.Errorf("SQLite database decompress: %s", err)
			}

			if err := gz.Close(); err != nil {
				return nil, err
			}
			database = buf.Bytes()
		} else {
			database = b[offset : offset+int64(sz)]
		}
	} else {
		database = nil
	}
	return database, nil
}

func GetNodeIPWithRaftAddress(logger *log.Logger, raftAddress string) string {
	configs := config.GetScheduler0Configurations(logger)

	for _, replica := range configs.Replicas {
		if replica.RaftAddress == raftAddress {
			return replica.Address
		}
	}

	return ""
}
