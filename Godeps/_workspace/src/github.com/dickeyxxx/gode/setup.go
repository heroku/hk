package gode

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// IsSetup returns true if node is setup in the client's RootPath directory
func (c *Client) IsSetup() bool {
	// TODO: better check if it is setup
	exists, _ := fileExists(c.NodePath)
	return exists
}

// Setup downloads and sets up node in the client's RootPath directory
func (c *Client) Setup() error {
	err := os.MkdirAll(c.RootPath, 0777)
	if err != nil {
		return err
	}
	resp, err := http.Get(c.NodeURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	uncompressed, err := gzip.NewReader(resp.Body)
	if err != nil {
		return err
	}
	return extractArchive(tar.NewReader(uncompressed), c.RootPath)
}

func extractArchive(archive *tar.Reader, rootPath string) error {
	for {
		hdr, err := archive.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		path := filepath.Join(rootPath, hdr.Name)
		switch {
		case hdr.FileInfo().IsDir():
			if err := os.Mkdir(path, hdr.FileInfo().Mode()); err != nil {
				return err
			}
		case hdr.Linkname != "":
			if err := os.Symlink(hdr.Linkname, path); err != nil {
				return err
			}
		default:
			if err := extractFile(archive, hdr, path); err != nil {
				return err
			}
		}
	}
}

func extractFile(archive *tar.Reader, hdr *tar.Header, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, archive)
	if err != nil {
		return err
	}
	return os.Chmod(path, hdr.FileInfo().Mode())
}
