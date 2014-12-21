package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	u "net/url"
)

func main() {
	// seedUrl, err := u.Parse("http://devotter.com")
	seedUrl, err := u.Parse("http://curtis.io")
	if err != nil {
		log.Fatal(err)
	}

	var crawledUrls = make(map[u.URL]bool)
	crawlResults := crawl(*seedUrl, crawledUrls)

	for link := range crawlResults {
		fmt.Printf("output: %v\n", link)
	}

	fmt.Println("done")
	return
}

type TestedUrl struct {
	url    u.URL
	status int
}

func FindLinks(url u.URL) (chan TestedUrl, chan u.URL, chan error) {
	links := make(chan u.URL)
	initResponse := make(chan TestedUrl)
	errChan := make(chan error)

	go func() {
		res, err := http.Get(url.String())
		if err != nil {
			errChan <- err
		}

		initResponse <- TestedUrl{*res.Request.URL, res.StatusCode}
		close(initResponse)

		doc, err := goquery.NewDocumentFromResponse(res)
		if err != nil {
			errChan <- err
		}

		doc.Find("a").Each(func(_ int, sel *goquery.Selection) {
			href, _ := sel.Attr("href")
			parsed, err := u.Parse(href)
			if err != nil {
				log.Fatal(err)
			}
			links <- *url.ResolveReference(parsed)
		})
		close(links)
	}()

	return initResponse, links, errChan
}

func crawl(seedUrl u.URL, crawled map[u.URL]bool) chan TestedUrl {

	urls := make(chan TestedUrl)

	res, links, errChan := FindLinks(seedUrl)
	crawled[seedUrl] = true
	urls <- <-res

	nextRound := make([]u.URL, 0)
	for link := range links {
		_, alreadyCrawled := crawled[link]
		if !alreadyCrawled {
			nextRound = append(nextRound, link)
		}
	}

	for {
		if len(nextRound) == 0 {
			break
		}

		linkChans := make([]chan u.URL, 0)
		for _, link := range nextRound {
			res, linkChan := FindLinks(link)
			crawled[link] = true
			urls <- <-res

			if isInternalLink := link.Host == seedUrl.Host; isInternalLink {
				linkChans = append(linkChans, linkChan)
			} else {
				go func() { urls <- <-test(link) }()
			}
		}

		fmt.Println("seen: ", crawled)
		nextRound = make([]u.URL, 0)
		for _, c := range linkChans {
			for link := range c {
				_, alreadyCrawled := crawled[link]
				if !alreadyCrawled {
					nextRound = append(nextRound, link)
				}
			}
		}
	}

	close(urls)

	return urls
}

func test(url u.URL) chan TestedUrl {
	c := make(chan TestedUrl)
	go func() {
		res, err := http.Head(url.String())
		if err != nil {
			// log.Fatal(err)
			// return nil, err
			u, _ := u.Parse("http://google.com")
			c <- TestedUrl{*u, 0}
		}
		c <- TestedUrl{url, res.StatusCode}
	}()
	return c
}
