package storage

import (
	"io"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

type FSStore struct {
	baseDir string
}

func NewFSStore(baseDir string) (*FSStore, error) {
	log.Info().Msgf("Saving file to %s", baseDir)
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		log.Error().Err(err).Msg("Failed to create base FS directories")
		return nil, err
	}
	return &FSStore{baseDir: baseDir}, nil
}
func (s *FSStore) BaseDir() string {
	return s.baseDir
}
func (s *FSStore) Save(path string, r io.Reader) error {
	// create dirs
	fullPath := filepath.Join(s.baseDir, path)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		log.Error().Err(err).Msg("Failed to create directory for file")
		return err
	}
	log.Info().Msgf("Saving file to %s", fullPath)

	// create file
	f, err := os.Create(fullPath)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create file")
		return err
	}
	defer f.Close()

	// write file
	_, err = io.Copy(f, r)

	return err
}
func (s *FSStore) Read(path string) (*os.File, error) {
	return os.Open(filepath.Join(s.baseDir, path))
}
func (s *FSStore) Delete(path string) error {
	return os.Remove(filepath.Join(s.baseDir, path))
}
func (s *FSStore) List(prefix string) ([]string, error) {
	var files []string
	root := filepath.Join(s.baseDir, prefix)
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		// on walk
		if err != nil {
			log.Error().Err(err).Msgf("Error walking through files %s", path)
			return err
		}
		if !info.IsDir() {
			rel, _ := filepath.Rel(s.baseDir, path)
			files = append(files, rel)
		}
		return nil
	})
	return files, err
}
