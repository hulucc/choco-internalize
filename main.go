package main

import (
	"flag"
	"fmt"
	"log"
)

func main() {
	id := flag.String("id", "", "choco pack id")
	ver := flag.String("ver", "", "choco pack ver")
	proxyPrefix := flag.String("proxy-prefix", "", "download cache proxy prefix")
	push := flag.Bool("push", false, "enable push")
	pushSource := flag.String("push-source", "", "push source")
	pushApiKey := flag.String("push-apikey", "", "push api key")
	keepTemp := flag.Bool("keep-temp", false, "keep temp")
	flag.Parse()
	if *id == "" || *ver == "" || *proxyPrefix == "" {
		log.Fatalln("[FATAL]", fmt.Errorf("required argment missing"))
	}
	if *push {
		if *pushSource == "" || *pushApiKey == "" {
			log.Fatalln("[FATAL]", fmt.Errorf("required push argment missing"))
		}
	}
	config := Config{
		DownloadCacheProxyPrefix: *proxyPrefix,
		KeepTemp:                 *keepTemp,
		Push:                     *push,
		PushSource:               *pushSource,
		PushApiKey:               *pushApiKey,
	}
	if err := ProcessPackage(config, *id, *ver); err != nil {
		log.Fatalln("[FATAL]", fmt.Errorf("ProcessPackage err: %w", err))
	}
}
