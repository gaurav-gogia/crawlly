package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/gocolly/colly"
	"golang.org/x/exp/slices"
)

const (
	findExp      = `^\w+([-+.']\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*$`
	leadCloseExp = `^[\s\p{Zs}]+|[\s\p{Zs}]+$`
	insidersExp  = `[\s\p{Zs}]{2,}`
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("need url as cli argument")
	}

	root := os.Args[1]

	c := colly.NewCollector()

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		if strings.HasPrefix(e.Attr("href"), "javascript") {
			return
		}
		e.Request.Visit(e.Attr("href"))
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
		page := fetch(r.URL.String())
		standard := standardspace(page)
		text := find(strings.Split(standard, " "))
		save("./data", text)
	})

	c.Visit(root)

	fmt.Println("END")
}

func fetch(url string) string {
	res, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func find(data []string) (found []string) {
	r := regexp.MustCompile(findExp)
	for i := range data {
		if r.MatchString(data[i]) {
			if slices.Contains(found, data[i]) {
				continue
			}
			found = append(found, data[i])
		}
	}
	return
}

func save(name string, text []string) {
	f, err := os.OpenFile(name, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	for t := range text {
		fmttext := strings.Trim(text[t], " ")
		f.WriteString(fmttext + "\n")
	}
}

func standardspace(page string) string {
	leadClose := regexp.MustCompile(leadCloseExp)
	insiders := regexp.MustCompile(insidersExp)
	final := leadClose.ReplaceAllString(page, "")
	return insiders.ReplaceAllString(final, " ")
}
