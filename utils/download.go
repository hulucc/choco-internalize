package utils

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func DownloadPackage(url string, w io.Writer) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("http.Get request %s err: %w", url, err)
	}
	defer resp.Body.Close()
	if _, err := io.Copy(w, resp.Body); err != nil {
		return fmt.Errorf("download err: %w", err)
	}
	return nil
}

func ValidateDownloadCache(url string) error {
	done := make(chan error)
	go func() {
		defer close(done)
		resp, err := http.Get(url)
		if err != nil {
			done <- fmt.Errorf("http.Get request %s err: %w", url, err)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			done <- fmt.Errorf("http.Get %s response err: code %d", url, resp.StatusCode)
			return
		}
		done <- nil
		return
	}()
	var err error
	select {
	case err = <-done:
	case <-time.After(time.Second):
		log.Println("[INFO]", fmt.Sprintf("Validating download cache %s", url))
		err = <-done
	}
	return err
}

func Unzip(zipFile *os.File, dest string) error {
	stat, err := zipFile.Stat()
	if err != nil {
		return fmt.Errorf("zipFile.Stat(%s) err: %w", zipFile.Name(), err)
	}
	r, err := zip.NewReader(zipFile, stat.Size())
	if err != nil {
		return fmt.Errorf("zip.NewReader(%s) err: %w", zipFile.Name(), err)
	}
	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("f.Open(%s) err: %w", f.Name, err)
		}
		defer rc.Close()

		fpath := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
		} else {
			var fdir string
			if lastIndex := strings.LastIndex(fpath, string(os.PathSeparator)); lastIndex > -1 {
				fdir = fpath[:lastIndex]
			}
			err = os.MkdirAll(fdir, os.ModePerm)
			if err != nil {
				return fmt.Errorf("os.MkdirAll(%s) err: %w", fdir, err)
			}
			f, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return fmt.Errorf("os.OpenFile(%s) err: %w", fpath, err)
			}
			defer f.Close()
			_, err = io.Copy(f, rc)
			if err != nil {
				return fmt.Errorf("io.Copy(%s) err: %w", fpath, err)
			}
		}
	}
	return nil
}

func UnpackPackageFromUrl(url string) (string, error) {
	out, err := os.CreateTemp("", "*.nupkg")
	if err != nil {
		return "", fmt.Errorf("os.CreateTemp")
	}
	defer os.Remove(out.Name())
	defer out.Close()
	if err := DownloadPackage(url, out); err != nil {
		return "", fmt.Errorf("DownloadPackage err: %w", err)
	}
	dir := strings.TrimSuffix(out.Name(), ".nupkg")
	os.MkdirAll(dir, 0777)
	if err := Unzip(out, dir); err != nil {
		return "", fmt.Errorf("UnzipPackage err: %w", err)
	}
	return dir, nil
}
