package storage

import "io"

type Store interface {
	Save(path string, r io.Reader) error
	Read(path string) (io.ReadCloser, error)
	Delete(path string) error
	List(prefix string) ([]string, error)
}
