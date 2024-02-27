package manifest

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

var _manifestParser *parser

type Parser interface {
	LoadManifest(path string) (*Manifest, error)
}

func GlobalManifestParser() Parser {
	if _manifestParser == nil {
		_manifestParser = &parser{}
	}

	return _manifestParser
}

func NewManifestParser() Parser {
	return &parser{}
}

type parser struct {
	isValid  bool
	manifest Manifest
	path     string
}

func (m *parser) LoadManifest(path string) (*Manifest, error) {
	filePath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("absoluting path: %w", err)
	}

	if m.isValid && path == filePath {
		return &m.manifest, nil
	}

	err = m.loadDbtFile(filePath)
	if err != nil {
		return nil, err
	}

	m.isValid = true
	path = filePath

	return &m.manifest, nil
}

func (m *parser) loadDbtFile(dbtFilePath string) error {
	jsonBytes, err := os.ReadFile(dbtFilePath)
	if err != nil {
		return fmt.Errorf("reading dbt file: %w", err)
	}

	err = json.Unmarshal(jsonBytes, &m.manifest)
	if err != nil {
		return fmt.Errorf("parsing dbt file: %w", err)
	}

	return nil
}
