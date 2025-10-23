package storage

import (
	"io"
	"os"
)

type Store interface {
	Save(path string, r io.Reader) error
	Read(path string) (*os.File, error)
	Delete(path string) error
	List(prefix string) ([]string, error)
	BaseDir() string
}
