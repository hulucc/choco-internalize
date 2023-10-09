package nuget

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
)

func Pack(ns *NuSpec, dir string, out io.Writer) error {
	// Assume filename from ID
	nsfilename := ns.Meta.ID + ".nuspec"
	// Fix file references to match current underlying os
	for x, f := range ns.Files.File {
		ns.Files.File[x].Source = strings.ReplaceAll(f.Source, `\`, string(os.PathSeparator))
		ns.Files.File[x].Target = strings.ReplaceAll(f.Target, `\`, string(os.PathSeparator))
	}
	// Create a new zip archive
	w := zip.NewWriter(out)
	defer w.Close()
	// Create a new Contenttypes Structure
	ct := NewContentTypes()
	// Add .nuspec to Archive
	b, err := ns.ToBytes()
	if err != nil {
		return fmt.Errorf("ns.ToBytes() err: %w", err)
	}
	if err := archiveFile(filepath.Base(nsfilename), w, b); err != nil {
		return fmt.Errorf("archiveFile err: %w", err)
	}
	ct.Add(filepath.Ext(nsfilename))
	// Process files
	// If there are no file globs specified then
	if len(ns.Files.File) == 0 {
		// walk the dir and zip up all found files. Everything.]
		err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() && filepath.Base(path) != filepath.Base(nsfilename) {
				// Open the file
				x, err := os.Open(path)
				if err != nil {
					return fmt.Errorf("os.Open err: %w", err)
				}
				// Gather all contents
				y, err := ioutil.ReadAll(x)
				if err != nil {
					return fmt.Errorf("ioutil.ReadAll err: %w", err)
				}
				// Set relative path for file in archive
				p, err := filepath.Rel(dir, path)
				if err != nil {
					return fmt.Errorf("filepath.Rel err: %w", err)
				}
				// Store the file
				if err := archiveFile(p, w, y); err != nil {
					return fmt.Errorf("archiveFile err: %w", err)
				}
				// Add extension to the Rels file
				ct.Add(filepath.Ext(p))
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("filepath.Walk err: %w", err)
		}
	} else {
		// For each of the specified globs, get files an put in target
		for _, f := range ns.Files.File {
			// Apply glob
			// ToDo: Fix Source from Windows to Current via os.seperator and strings replace
			matches, err := filepath.Glob(filepath.Join(dir, f.Source))
			if err != nil {
				return fmt.Errorf("filepath.Glob err: %w", err)
			}
			for _, m := range matches {
				info, err := os.Stat(m)
				if !info.IsDir() && filepath.Base(m) != filepath.Base(nsfilename) {
					// Open the file
					x, err := os.Open(m)
					if err != nil {
						return fmt.Errorf("os.Open err: %w", err)
					}
					// Gather all contents
					y, err := ioutil.ReadAll(x)
					if err != nil {
						return fmt.Errorf("ioutil.ReadAll err: %w", err)
					}
					// Set relative path for file in archive
					p, err := filepath.Rel(dir, m)
					if err != nil {
						return fmt.Errorf("filepath.Rel err: %w", err)
					}
					// Overide path if Target is set
					if f.Target != "" {
						p = filepath.Join(f.Target, filepath.Base(m))
					}
					// Store the file
					if err := archiveFile(p, w, y); err != nil {
						return fmt.Errorf("archiveFile err: %w", err)
					}
					// Add extension to the Rels file
					ct.Add(filepath.Ext(p))
				}
				if err != nil {
					return fmt.Errorf("os.Stat err: %w", err)
				}
			}
		}
	}

	// Create and add .psmdcp file to Archive
	pf := NewPsmdcpFile()
	pf.Creator = ns.Meta.Authors
	pf.Description = ns.Meta.Description
	pf.Identifier = ns.Meta.ID
	pf.Version = ns.Meta.Version
	pf.Keywords = ns.Meta.Tags
	pf.LastModifiedBy = "go-nuget"
	b, err = pf.ToBytes()
	if err != nil {
		return fmt.Errorf("pf.ToBytes() err: %w", err)
	}
	pfn := "package/services/metadata/core-properties/" + randomString(32) + ".psmdcp"
	if err := archiveFile(pfn, w, b); err != nil {
		return fmt.Errorf("archiveFile err: %w", err)
	}
	ct.Add(filepath.Ext(pfn))

	// Create and add .rels to Archive
	rf := NewRelFile()
	rf.Add("http://schemas.microsoft.com/packaging/2010/07/manifest", "/"+filepath.Base(nsfilename))
	rf.Add("http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties", pfn)

	b, err = rf.ToBytes()
	if err != nil {
		return fmt.Errorf("rf.ToBytes() err: %w", err)
	}
	if err := archiveFile(filepath.Join("_rels", ".rels"), w, b); err != nil {
		return fmt.Errorf("archiveFile err: %w", err)
	}
	ct.Add(filepath.Ext(".rels"))

	// Add [Content_Types].xml to Archive
	b, err = ct.ToBytes()
	if err != nil {
		return fmt.Errorf("ct.ToBytes() err: %w", err)
	}
	if err := archiveFile(`[Content_Types].xml`, w, b); err != nil {
		return fmt.Errorf("archiveFile err: %w", err)
	}

	return nil
}

func archiveFile(fn string, w *zip.Writer, b []byte) error {

	// Create the file in the zip
	f, err := w.Create(fn)
	if err != nil {
		return fmt.Errorf("w.Create err: %w", err)
	}

	// Write .nuspec bytes to file
	_, err = f.Write([]byte(b))
	if err != nil {
		return fmt.Errorf("f.Write err: %w", err)
	}
	return nil
}

const letterBytes = "abcdef0123456789"

func randomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
