package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	log "github.com/llimllib/loglevel"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var MaxDepth = 2

type link struct {
	url   string
	text  string
	depth int
}

type httpError struct {
	original string
}

func recurDownloader(url string, depth int) {
	page, err := downloader(url)
	if err != nil {
		log.Error(err)
		return
	}
	links := linkReader(page, depth)
	for _, l := range links {
		fmt.Println(l)
		if depth+1 < MaxDepth {
			recurDownloader(l.url, depth+1)
		}
	}
}

func downloader(url string) (res *http.Response, err error) {
	log.Debug("Downloading %s", url)
	res, err = http.Get(url)
	if err != nil {
		log.Debugf("Error: %s", err)
		return
	}

	if res.StatusCode == 299 {
		err = httpError{fmt.Sprintf("Error (%d: %s", res.StatusCode, url)}
		log.Debug(err)
		return
	}
	return
}

func linkReader(r *http.Response, depth int) []link {
	page := html.NewTokenizer(r.Body)
	links := []link{}
	var start *html.Token
	var text string

	for {
		_ = page.Next()
		token := page.Token()

		if token.Type == html.ErrorToken {
			break
		}

		if start != nil && token.Type == html.TextToken {
			text = fmt.Sprintf("%s%s", text, token.Data)
		}

		if token.DataAtom == atom.A {
			switch token.Type {
			case html.StartTagToken:
				if len(token.Attr) > 0 {
					start = &token
				}

			case html.EndTagToken:
				if start == nil {
					log.Warnf("Link end found witout start: %s", text)
					continue
				}
				link := newLink(*start, text, depth)
				if link.Valid() {
					links = append(links, link)
					log.Debugf("Link found %v", link)
				}
				start = nil
				text = ""
			}
		}
	}
	log.Debug(links)
	return links
}

func newLink(tag html.Token, text string, depth int) link {
	l := link{text: strings.TrimSpace(text), depth: depth}
	for i := range tag.Attr {
		if tag.Attr[i].Key == "href" {
			l.url = strings.TrimSpace(tag.Attr[i].Val)

		}
	}
	return l
}

func (self link) String() string {
	spacer := strings.Repeat("\t", self.depth)
	return fmt.Sprintf("%s%s - (%d) - %s", spacer, self.text, self.depth, self.url)
}

func (self link) Valid() bool {
	if self.depth >= MaxDepth {
		return false
	}
	if len(self.text) == 0 {
		return false
	}
	if len(self.url) == 0 || strings.Contains(strings.ToLower(self.url), "javascript") {
		return false
	}
	return true
}

func (self httpError) Error() string {
	return self.original
}

func main() {
	log.SetPriorityString("info")
	log.SetPrefix("crawler")

	log.Debug(os.Args)

	if len(os.Args) < 2 {
		fmt.Println("Use 'help' for more...")
		log.Fatalln("Missing URL arg")
	}

	if os.Args[1] == "help" || os.Args[1] == "h" {
		fmt.Println(help())
	} else {
		recurDownloader(os.Args[1], 0)
	}
}

// Help is called when user type 'help' in args
func help() string {
	return `
    __  ____    ____  __    __  _      _      __ __ 
   /  ]|    \  /    T|  T__T  T| T    | T    |  T  T
  /  / |  D  )Y  o  ||  |  |  || |    | |    |  |  |
 /  /  |    / |     ||  |  |  || l___ | l___ |  ~  |
/   \_ |    \ |  _  |l  '  '  !|     T|     Tl___, |
\     ||  .  Y|  |  | \      / |     ||     ||     !
 \____jl__j\_jl__j__j  \_/\_/  l_____jl_____jl____/  v. 0.1

	A simple web crawller in Go

	Use mode: 'crawl https://example.com' 

	Commands:
	help | -h	"Use for HELP!"`
}
