package crawl

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	u "net/url"
	str "strings"
)

type TestedUrl struct {
	Url        u.URL
	Status     int
	LinkedUrls []u.URL
}

func stripInPageLink(s string) string {
	lastIndex := str.LastIndex(s, "#")
	if lastIndex != -1 {
		s = s[0:lastIndex]
	}
	return s
}

func FindLinks(url u.URL) (chan TestedUrl, chan error) {
	c := make(chan TestedUrl)
	errChan := make(chan error)

	go func() {
		res, err := http.Get(url.String())
		if err != nil {
			errChan <- err
		}

		doc, err := goquery.NewDocumentFromResponse(res)
		if err != nil {
			errChan <- err
		}

		linkedUrls := make([]u.URL, 0)
		doc.Find("a").Each(func(_ int, sel *goquery.Selection) {
			href, _ := sel.Attr("href")
			parsed, err := u.Parse(stripInPageLink(href))
			if err != nil {
				fmt.Println("failed to parse: ", href)
				errChan <- err
				fmt.Println("done with error handling")
				return
			}
			linkedUrls = append(linkedUrls, *url.ResolveReference(parsed))
		})
		c <- TestedUrl{url, res.StatusCode, linkedUrls}
		close(c)
	}()

	return c, errChan
}

func Crawl(seedUrl u.URL) chan TestedUrl {
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
		toTest = tested.LinkedUrls

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
					if shouldCrawl(seedUrl, testedChild.Url) {
						fmt.Println("Crawling children of: ", testedChild.Url)
						nextRound = append(nextRound, testedChild.LinkedUrls...)
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
