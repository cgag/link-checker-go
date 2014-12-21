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

func FindLinks(url u.URL) (chan TestedUrl, chan u.URL) {
	links := make(chan u.URL)
	initResponse := make(chan TestedUrl)

	go func() {
		res, err := http.Get(url.String())
		if err != nil {
			log.Fatal(err)
		}
		initResponse <- TestedUrl{*res.Request.URL, res.StatusCode}
		close(initResponse)

		doc, err := goquery.NewDocumentFromResponse(res)
		if err != nil {
			log.Fatal(err)
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

	return initResponse, links
}

func crawl(seedUrl u.URL, crawled map[u.URL]bool) chan TestedUrl {

	urls := make(chan TestedUrl)
	go func() {
		res, links := FindLinks(seedUrl)

		seedRes := <-res
		crawled[seedUrl] = true
		urls <- seedRes

		// toTest := make(chan u.URL)
		// testResults := make(chan TestedUrl)

		// go func() {
		// 	for link := range toTest {
		// 		testResults <- test(link)
		// 	}
		// }()

		for link := range links {
			_, alreadyCrawled := crawled[link]
			if !alreadyCrawled {
				if isInternalLink := link.Host == seedUrl.Host; isInternalLink {
					for u := range crawl(link, crawled) {
						urls <- u
					}
				} else {
					// internal links get checked by the recursive call to crawl
					crawled[link] = true
					urls <- test(link)
					// toTest <- link
				}
			}
		}
		// close(toTest)
		// fmt.Println("closed totest")

		// for tested := range testResults {
		// 	urls <- tested
		// }
		close(urls)
	}()

	return urls
}

func test(url u.URL) TestedUrl {
	res, err := http.Head(url.String())
	if err != nil {
		log.Fatal(err)
	}
	return TestedUrl{url, res.StatusCode}
}
