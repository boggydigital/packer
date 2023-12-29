package packer

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"github.com/boggydigital/nod"
	"io"
	"os"
	"path/filepath"
	"time"
)

func Pack(from, to string, tpw nod.TotalProgressWriter) error {

	if from == "" || to == "" {
		return errors.New("packing requires from and to dirs")
	}

	root, _ := filepath.Split(from)

	efn := fmt.Sprintf(
		"%s.tar.gz",
		time.Now().Format(nod.TimeFormat))

	exportedPath := filepath.Join(to, efn)

	if _, err := os.Stat(exportedPath); os.IsExist(err) {
		return err
	}

	file, err := os.Create(exportedPath)
	if err != nil {
		return err
	}
	defer file.Close()

	gw := gzip.NewWriter(file)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	files := make([]string, 0)

	if err := filepath.Walk(from, func(f string, fi os.FileInfo, err error) error {
		if fi.IsDir() {
			return nil
		}
		files = append(files, f)
		return nil
	}); err != nil {
		return err
	}

	if tpw != nil {
		tpw.TotalInt(len(files))
	}

	for _, f := range files {

		fi, err := os.Stat(f)
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(fi, f)
		if err != nil {
			return err
		}

		rp, err := filepath.Rel(root, f)
		if err != nil {
			return err
		}

		header.Name = filepath.ToSlash(rp)

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		of, err := os.Open(f)
		if err != nil {
			return err
		}

		if _, err := io.Copy(tw, of); err != nil {
			return err
		}

		if tpw != nil {
			tpw.Increment()
		}
	}

	return nil
}
