package service

import (
	"fmt"
	"html/template"
	"net/http"
)

type newsPage struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	Month    string `json:"month"`
	Day      string `json:"day"`
	Links    []Link `json:"links"`
	Keyword  string `json:"keyword"`
}

func newsPageHandle(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("keyword")
	fmt.Printf("[keyword] ==> \"%s\"\n", keyword)
	month := r.URL.Query().Get("month")
	day := r.URL.Query().Get("day")
	choice := r.URL.Query().Get("nc")

	// read local files
	links, err := readLinksFromJSON(GetRawNewsPath(keyword))
	if err != nil {
		// if local files are already deleted, crawl again
		_, err := NewCrawlerWithKeyword(keyword)
		if err != nil {
			returnErr(err, w)
			return
		}
		links, err = readLinksFromJSON(GetRawNewsPath(keyword))
		if err != nil {
			returnErr(err, w)
			return
		}
	}

	// render news page
	if len(links) != 0 {
		tmpl := template.Must(template.ParseFiles("templates/news.html"))
		if keyword == "" {
			links, err := chooseLinksFromKey(links, choice)
			if err != nil {
				returnErr(err, w)
				return
			}
			data := newsPage{
				Title:    "news",
				Subtitle: fmt.Sprintf("%s/%s", month, day),
				Month:    month,
				Day:      day,
				Links:    links,
				Keyword:  "",
			}
			tmpl.Execute(w, data)
		} else {
			data := newsPage{
				Title:    fmt.Sprintf("news about %s", keyword),
				Subtitle: fmt.Sprintf("%s/%s", month, day),
				Month:    month,
				Day:      day,
				Links:    links,
				Keyword:  keyword,
			}
			tmpl.Execute(w, data)
		}
	} else {
		returnErr(newsHandleError{"Links not found"}, w)
		return
	}
}
