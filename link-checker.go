package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	u "net/url"
)

func main() {
	seedUrl, err := u.Parse("http://devotter.com")
	if err != nil {
		err.Error()
	}

	var chans = make([]chan *u.URL, 0)
	var crawledUrls = make(map[u.URL]bool)

	links := FindLinks(seedUrl)
	crawledUrls[*seedUrl] = true

	for link := range links {
		fmt.Printf("link: %v\n", link)
		crawledUrls[*link] = true
		if isInternalLink := link.Host == seedUrl.Host; isInternalLink {
			moreLinks := FindLinks(link)
			chans = append(chans, moreLinks)
		}
	}

	for _, c := range chans {
		for link := range c {
			_, alreadyCrawled := crawledUrls[*link]
			if !alreadyCrawled {
				fmt.Printf("link: %v\n", link)
			}
		}
	}

	fmt.Println("done")
	return
}

func FindLinks(url *u.URL) chan *u.URL {
	c := make(chan *u.URL)

	go func() {
		doc, _ := goquery.NewDocument(url.String())
		doc.Find("a").Each(func(_ int, sel *goquery.Selection) {
			href, _ := sel.Attr("href")
			parsed, err := u.Parse(href)
			if err != nil { // if only we had Maybe
				err.Error()
			}
			c <- url.ResolveReference(parsed)
		})
		close(c)
	}()

	return c
}

//

// if IsRelative(parsed) {
// 					parsed = url.ResolveReference(parsed)
// 				}
// 				_, alreadyCrawled := crawled[*parsed]
// 				if !alreadyCrawled {
// 					c <- parsed
// 					crawled[*parsed] = true // bad name, we dont really crawl external links

// 					if parsed.Host == url.Host {
// 						link, d := crawl(parsed, crawled)
// 						for {
// 							select {
// 							case l := <-link:
// 								c <- l
// 							case <-d:
// 								done <- "done"
// 							}
// 						}
// 					}
// 				}
// 				done <- "done"
