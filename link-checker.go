package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	u "net/url"
	// "time"
)

// var totalTime = time.ParseDuration("0ms")

func main() {
	// seedUrl, err := u.Parse("http://devotter.com")
	// seedUrl, err := u.Parse("http://curtis.io")
	// seedUrl, err := u.Parse("http://me.justin.sh")
	seedUrl, err := u.Parse("https://ocharles.org.uk")
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

func FindLinks(url u.URL) (chan TestedUrl, chan error) {
	c := make(chan TestedUrl)
	errChan := make(chan error)

	go func() {
		// t1 := time.Now()
		res, err := http.Get(url.String())
		if err != nil {
			errChan <- err
		}
		// delta := time.Since(t1)
		// totalTime += delta
		// fmt.Println("fetched in: ", delta)

		doc, err := goquery.NewDocumentFromResponse(res)
		if err != nil {
			errChan <- err
		}

		linkedUrls := make([]u.URL, 0)
		doc.Find("a").Each(func(_ int, sel *goquery.Selection) {
			href, _ := sel.Attr("href")
			parsed, err := u.Parse(href)
			if err != nil {
				errChan <- err
			}
			linkedUrls = append(linkedUrls, *url.ResolveReference(parsed))
		})
		c <- TestedUrl{url, res.StatusCode, linkedUrls}
		close(c)
	}()

	return c, errChan
}

func crawl(seedUrl u.URL) chan TestedUrl {
	retUrls := make(chan TestedUrl)
	go func() {
		crawled := make(map[u.URL]bool)

		testedChan, errChan := FindLinks(seedUrl)

		var tested TestedUrl
		select {
		case t := <-testedChan:
			tested = t
			break
		case err := <-errChan:
			log.Fatal(err)
		}

		crawled[seedUrl] = true
		retUrls <- tested

		toTest := make([]u.URL, 0)
		toTest = tested.linkedUrls

		for {
			nextRound := make([]u.URL, 0)

			childChans := make([]chan TestedUrl, 0)
			errChans := make([]chan error, 0)

			for _, child := range toTest {
				if _, alreadyCrawled := crawled[child]; !alreadyCrawled {
					crawled[child] = true
					testedChildChan, errChan := FindLinks(child)
					childChans = append(childChans, testedChildChan)
					errChans = append(errChans, errChan)
				}
			}
			fmt.Println("done making chans: ", len(childChans))

			for i := range childChans {
				c, ec := childChans[i], errChans[i]

				select {
				case testedChild := <-c:
					retUrls <- testedChild
					if shouldCrawl(seedUrl, testedChild.url) {
						nextRound = append(nextRound, testedChild.linkedUrls...)
					}
				case <-ec:
					continue
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
