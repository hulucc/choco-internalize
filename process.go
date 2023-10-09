package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/hulucc/choco-internalize/choco"
	"github.com/hulucc/choco-internalize/nuget"
	"github.com/hulucc/choco-internalize/utils"
)

type Config struct {
	DownloadCacheProxyPrefix string
	KeepTemp                 bool
	Push                     bool
	PushSource               string
	PushApiKey               string
}

func Internalize(config Config, md *choco.PkgMetadata, dir string) error {
	path := filepath.Join(dir, "tools/ChocolateyInstall.ps1")
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			log.Println("[INFO]", fmt.Sprintf("%s is already internal", md))
			return nil
		}
		return fmt.Errorf("os.Stat(%s) err: %w", path, err)
	}
	cache, err := md.ParseDownloadCache()
	if err != nil {
		return fmt.Errorf("ParseDownloadCache err: %w", err)
	}
	if len(cache) == 0 {
		log.Println("[INFO]", fmt.Sprintf("%s had no download content to internalize", md))
		return nil
	}

	bs, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("ioutil.ReadFile(%s) err: %w", path, err)
	}
	content := string(bs)
	for _, item := range cache {
		target := strings.Replace(item.Url, "https://", "https/", 1)
		target = strings.Replace(target, "http://", "http/", 1)
		target = config.DownloadCacheProxyPrefix + target
		if err := utils.ValidateDownloadCache(target); err != nil {
			log.Println("[ERROR]", fmt.Errorf("%s download cache %s validate err: %w", md, target, err))
		} else {
			log.Println("[INFO]", fmt.Sprintf("%s download cache %s validated", md, target))
		}
		temp := strings.ReplaceAll(content, item.Url, target)
		if temp == content {
			if strings.Contains(content, target) {
				log.Println("[WARN]", fmt.Errorf("%s download cache %s already updated", md, target))
			} else {
				log.Println("[ERROR]", fmt.Errorf("%s download cache %s not found", md, item.Url))
			}
		} else {
			log.Println("[INFO]", fmt.Sprintf("%s download cache updated %s => %s", md, item.Url, target))
			content = temp
		}
	}
	if err := ioutil.WriteFile(path, []byte(content), 0777); err != nil {
		return fmt.Errorf("ioutil.WriteFile(%s) err: %w", path, err)
	}
	return nil
}

func ProcessPackage(config Config, id string, ver string) error {
	md, err := choco.GetPkgMetadata(id, ver)
	if err != nil {
		return fmt.Errorf("choco.GetPkgMetadata err: %w", err)
	}
	log.Println("[INFO]", fmt.Sprintf("downloading %s", md))
	dir, err := utils.UnpackPackageFromUrl(md.Content.Src)
	if err != nil {
		return fmt.Errorf("utils.UnpackPackageFromUrl(%s) err: %w", md.Content.Src, err)
	}
	log.Println("[INFO]", fmt.Sprintf("downloaded %s to %s", md, dir))
	if !config.KeepTemp {
		defer os.RemoveAll(dir)
	}
	if err := Internalize(config, md, dir); err != nil {
		return fmt.Errorf("Internalize err: %w", err)
	}
	specPath := filepath.Join(dir, fmt.Sprintf("%s.nuspec", md.Title))
	spec, err := nuget.SpecFromFile(specPath)
	if err != nil {
		return fmt.Errorf("nuget.SpecFromFile err: %w", err)
	}
	nupkgDir, err := os.MkdirTemp("", "*")
	if err != nil {
		return fmt.Errorf("os.MkdirTemp err: %w", err)
	}
	if !config.KeepTemp {
		defer os.RemoveAll(nupkgDir)
	} else {
		log.Println("[INFO]", fmt.Sprintf("repacking package to temp dir: %s", nupkgDir))
	}
	nupkgPath := filepath.Join(nupkgDir, fmt.Sprintf("%s.nupkg", md.Title))
	nupkg, err := os.Create(nupkgPath)
	if err != nil {
		return fmt.Errorf("os.Create(%s) err: %w", nupkgPath, err)
	}
	defer nupkg.Close()
	if err := nuget.Pack(spec, dir, nupkg); err != nil {
		return fmt.Errorf("nuget.Pack err: %w", err)
	}
	if config.Push {
		nupkg.Seek(0, 0)
		log.Println("[INFO]", fmt.Sprintf("Pushing %s to %s", md, config.PushSource))
		if err := nuget.Push(config.PushSource, config.PushApiKey, nupkg); err != nil {
			return fmt.Errorf("nuget.Push err: %w", err)
		}
		log.Println("[INFO]", fmt.Sprintf("Pushed %s to %s", md, config.PushSource))
	}

	return nil
}
