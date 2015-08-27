// Copyright (C) 2014 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at http://mozilla.org/MPL/2.0/.

// Package upgrade downloads and compares releases, and upgrades the running binary.
package upgrade

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/calmh/upgrade/signature"
	"github.com/kardianos/osext"
)

func ToURL(url string, key []byte) error {
	path, err := osext.Executable()
	if err != nil {
		return err
	}

	return upgradeToURL(path, url, key)
}

// Upgrade to the given release, saving the previous binary with a ".old" extension.
func upgradeToURL(binary, url string, key []byte) error {
	fname, sig, err := readRelease(filepath.Dir(binary), url)
	if err != nil {
		return err
	}

	if err := verifyUpgrade(fname, sig, key); err != nil {
		os.Remove(fname)
		return err
	}

	old := binary + ".old"
	os.Remove(old)
	err = os.Rename(binary, old)
	if err != nil {
		return err
	}
	err = os.Rename(fname, binary)
	if err != nil {
		return err
	}
	return nil
}

func readRelease(dir, url string) (string, []byte, error) {
	if debug {
		l.Debugf("loading %q", url)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", nil, err
	}

	req.Header.Add("Accept", "application/octet-stream")
	resp, err := insecureHTTP.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	switch runtime.GOOS {
	case "windows":
		return readZip(dir, resp.Body)
	default:
		return readTarGz(dir, resp.Body)
	}
}

func readTarGz(dir string, r io.Reader) (string, []byte, error) {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return "", nil, err
	}

	tr := tar.NewReader(gr)

	var tempName string
	var sig []byte

	// Iterate through the files in the archive.
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			// end of tar archive
			break
		}
		if err != nil {
			return "", nil, err
		}

		shortName := path.Base(hdr.Name)

		if debug {
			l.Debugf("considering file %q", shortName)
		}

		err = archiveFileVisitor(dir, &tempName, &sig, shortName, tr)
		if err != nil {
			return "", nil, err
		}

		if tempName != "" && sig != nil {
			break
		}
	}

	return tempName, sig, nil
}

func readZip(dir string, r io.Reader) (string, []byte, error) {
	body, err := ioutil.ReadAll(r)
	if err != nil {
		return "", nil, err
	}

	archive, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return "", nil, err
	}

	var tempName string
	var sig []byte

	// Iterate through the files in the archive.
	for _, file := range archive.File {
		shortName := path.Base(file.Name)

		if debug {
			l.Debugf("considering file %q", shortName)
		}

		inFile, err := file.Open()
		if err != nil {
			return "", nil, err
		}

		err = archiveFileVisitor(dir, &tempName, &sig, shortName, inFile)
		inFile.Close()
		if err != nil {
			return "", nil, err
		}

		if tempName != "" && sig != nil {
			break
		}
	}

	return tempName, sig, nil
}

// archiveFileVisitor is called for each file in an archive. It may set
// tempFile and signature.
func archiveFileVisitor(dir string, tempFile *string, signature *[]byte, filename string, filedata io.Reader) error {
	var err error
	switch filename {
	case "syncthing", "syncthing.exe":
		if debug {
			l.Debugln("reading binary")
		}
		*tempFile, err = writeBinary(dir, filedata)
		if err != nil {
			return err
		}

	case "syncthing.sig", "syncthing.exe.sig":
		if debug {
			l.Debugln("reading signature")
		}
		*signature, err = ioutil.ReadAll(filedata)
		if err != nil {
			return err
		}
	}

	return nil
}

func verifyUpgrade(tempName string, sig []byte, key []byte) error {
	if tempName == "" {
		return fmt.Errorf("no upgrade found")
	}
	if sig == nil {
		return fmt.Errorf("no signature found")
	}

	if debug {
		l.Debugf("checking signature\n%s", sig)
	}

	fd, err := os.Open(tempName)
	if err != nil {
		return err
	}
	err = signature.Verify(key, sig, fd)
	fd.Close()

	if err != nil {
		os.Remove(tempName)
		return err
	}

	return nil
}

func writeBinary(dir string, inFile io.Reader) (filename string, err error) {
	// Write the binary to a temporary file.

	outFile, err := ioutil.TempFile(dir, "syncthing")
	if err != nil {
		return "", err
	}

	_, err = io.Copy(outFile, inFile)
	if err != nil {
		os.Remove(outFile.Name())
		return "", err
	}

	err = outFile.Close()
	if err != nil {
		os.Remove(outFile.Name())
		return "", err
	}

	err = os.Chmod(outFile.Name(), os.FileMode(0755))
	if err != nil {
		os.Remove(outFile.Name())
		return "", err
	}

	return outFile.Name(), nil
}
