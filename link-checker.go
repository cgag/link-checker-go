package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	u "net/url"
)

func main() {
	seedUrl, err := u.Parse("http://devotter.com")
	// seedUrl, err := u.Parse("http://curtis.io")
	if err != nil {
		log.Fatal(err)
	}

	crawlResults := crawl(*seedUrl)

	for link := range crawlResults {
		fmt.Println("output: ", link.url, link.status)
	}

	fmt.Println("done")
	return
}

type TestedUrl struct {
	url        u.URL
	status     int
	linkedUrls []u.URL
}

func FindLinks(url u.URL) (TestedUrl, error) {
	res, err := http.Get(url.String())
	if err != nil {
		return TestedUrl{}, err
	}

	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		return TestedUrl{}, err
	}

	linkedUrls := make([]u.URL, 0)
	doc.Find("a").Each(func(_ int, sel *goquery.Selection) {
		href, _ := sel.Attr("href")
		parsed, err := u.Parse(href)
		if err != nil {
			return
		}
		linkedUrls = append(linkedUrls, *url.ResolveReference(parsed))
	})
	return TestedUrl{url, res.StatusCode, linkedUrls}, nil
}

func crawl(seedUrl u.URL) chan TestedUrl {
	retUrls := make(chan TestedUrl)
	go func() {
		crawled := make(map[u.URL]bool)

		fmt.Println("fetching first url")
		tested, err := FindLinks(seedUrl)
		fmt.Println("fetched first url")
		if err != nil {
			log.Fatal(err)
		}
		crawled[seedUrl] = true
		fmt.Println("pushing to output")
		retUrls <- tested

		toTest := make([]u.URL, 0)
		toTest = tested.linkedUrls

		for {
			fmt.Println("in crawl loop")
			nextRound := make([]u.URL, 0)
			for _, child := range toTest {
				if _, alreadyCrawled := crawled[child]; !alreadyCrawled {
					testedChild, err := FindLinks(child)
					if err != nil {
						continue
					}
					crawled[child] = true
					retUrls <- testedChild
					if shouldCrawl(seedUrl, child) {
						nextRound = append(nextRound, testedChild.linkedUrls...)
					}
				}
			}
			if len(nextRound) == 0 {
				break
			}
			toTest = nextRound
		}
		close(retUrls)
	}()

	return retUrls
}

func shouldCrawl(seedUrl u.URL, potential u.URL) bool {
	return seedUrl.Host == potential.Host
}
