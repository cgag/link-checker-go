package main

import (
	"curtis.io/link-checker/crawl"
	"fmt"
	"log"
	u "net/url"
)

func main() {
	seedUrl, err := u.Parse("http://devotter.com")
	if err != nil {
		log.Fatal(err)
	}

	crawlResults := crawl.Crawl(*seedUrl)

	for link := range crawlResults {
		fmt.Println("output: ", link.Url, link.Status)
	}

	fmt.Println("done")
	return
}
