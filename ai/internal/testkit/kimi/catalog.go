// Package kimi holds independent K3 qualification candidates. Availability is
// not formal support: each market/protocol must earn its own live acceptance.
package kimi

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"embed"
	"encoding/json"
	"fmt"
	"io"

	"github.com/tokenbeat-lab/barness/ai"
)

// Repeated literal records are intentional: protocols and markets evolve
// independently; no runtime model or price inherits a sibling identity.
//
//go:embed *.json sources/*
var files embed.FS

// CandidateCatalog returns fresh source-backed identities, never a promotion.
func CandidateCatalog() ai.Catalog {
	if err := VerifySources(); err != nil {
		panic(err)
	}
	var cat ai.Catalog
	raw, _ := files.ReadFile("candidate.json")
	if err := json.Unmarshal(raw, &cat); err != nil {
		panic(err)
	}
	if _, err := cat.Hash(); err != nil {
		panic(err)
	}
	return cat
}

// ProtocolCatalog narrows the candidate snapshot without modifying its source.
func ProtocolCatalog(provider ai.ProviderID, api ai.API) ai.Catalog {
	cat := CandidateCatalog()
	cat.Models = cat.ModelsOf(provider, api)
	return cat
}

// VerifySources verifies both retained transport bytes and complete originals.
// Raw sources are bounded by their pinned length and are never network-fetched.
func VerifySources() error {
	raw, err := files.ReadFile("source-files.json")
	if err != nil {
		return err
	}
	var index struct {
		Sources []struct {
			Path       string `json:"path"`
			SHA256     string `json:"sha256"`
			Compressed string `json:"compressedSHA256"`
			Bytes      int64  `json:"bytes"`
		} `json:"sources"`
	}
	if err := json.Unmarshal(raw, &index); err != nil {
		return err
	}
	for _, s := range index.Sources {
		packed, err := files.ReadFile("sources/" + s.Path)
		if err != nil {
			return err
		}
		if fmt.Sprintf("%x", sha256.Sum256(packed)) != s.Compressed {
			return fmt.Errorf("kimi: compressed source hash mismatch")
		}
		reader, err := gzip.NewReader(bytes.NewReader(packed))
		if err != nil {
			return err
		}
		original, err := io.ReadAll(io.LimitReader(reader, s.Bytes+1))
		closeErr := reader.Close()
		if err != nil {
			return err
		}
		if closeErr != nil {
			return closeErr
		}
		if int64(len(original)) != s.Bytes || fmt.Sprintf("%x", sha256.Sum256(original)) != s.SHA256 {
			return fmt.Errorf("kimi: original source mismatch")
		}
	}
	return nil
}
