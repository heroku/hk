package plugins

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/dickeyxxx/gonpm/cli"
)

func Setup() {
	if exists, _ := fileExists(nodePath); exists == true {
		return
	}
	cli.Stderrf("Setting up plugins... ")
	cli.Logln("Creating plugins directory")
	err := os.MkdirAll(cli.AppDir, 0777)
	must(err)
	cli.Logln("Downloading node from", NODE_URL)
	resp, err := http.Get(NODE_URL)
	must(err)
	defer resp.Body.Close()
	uncompressed, err := gzip.NewReader(resp.Body)
	must(err)
	cli.Logln("Extracting node to", nodePath)
	archive := tar.NewReader(uncompressed)
	for {
		hdr, err := archive.Next()
		if err == io.EOF {
			break
		}
		must(err)
		path := filepath.Join(cli.AppDir, hdr.Name)
		switch {
		case hdr.FileInfo().IsDir():
			err = os.Mkdir(path, 0777)
			must(err)
		case hdr.Linkname != "":
			err = os.Symlink(hdr.Linkname, path)
			must(err)
		default:
			file, err := os.Create(path)
			must(err)
			defer file.Close()
			_, err = io.Copy(file, archive)
			must(err)
		}
		err = os.Chmod(path, hdr.FileInfo().Mode())
		must(err)
	}
	cli.Logln("Finished installing node")
	cli.Stderrln("done")
}
