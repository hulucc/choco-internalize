package choco

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type PkgDownloadCache struct {
	Url string
	Path string
	Checksum string
}

type PkgMetadata struct {
	Title   string `xml:"title"`
	Content struct {
		Type string `xml:"type,attr"`
		Src  string `xml:"src,attr"`
	} `xml:"content"`
	Properties struct {
		Version string `xml:"http://schemas.microsoft.com/ado/2007/08/dataservices Version"`
		DownloadCache string `xml:"http://schemas.microsoft.com/ado/2007/08/dataservices DownloadCache"`
	} `xml:"http://schemas.microsoft.com/ado/2007/08/dataservices/metadata properties"`
}

func (it PkgMetadata) ParseDownloadCache() ([]PkgDownloadCache, error) {
	items := strings.Split(it.Properties.DownloadCache, "|")
	results := make([]PkgDownloadCache, 0, len(items))
	for _, item := range items {
		fields := strings.Split(item, "^")
		if len(fields) != 3 {
			return nil, fmt.Errorf("unknown download cache: %s", item)
		}
		result := PkgDownloadCache{Url: fields[0], Path: fields[1], Checksum: fields[2]}
		results = append(results, result)
	}
	return results, nil
}

func (it PkgMetadata) String() string {
	return fmt.Sprintf("%s.%s", it.Title, it.Properties.Version)
}

func GetPkgMetadata(id string, ver string) (*PkgMetadata, error) {
	url := fmt.Sprintf("https://community.chocolatey.org/api/v2/Packages(Id='%s',Version='%s')", id, ver)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("http.Get request err: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("http.Get %s response %d", url, resp.StatusCode)
	}
	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("io.ReadAll err: %w", err)
	}
	var md PkgMetadata
	if err := xml.Unmarshal(bs, &md); err != nil {
		return nil, fmt.Errorf("xml.Unmarshal err: %w", err)
	}
	return &md, nil
}
