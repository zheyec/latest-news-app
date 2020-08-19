package service

import (
	"fmt"
	"net/http"
	"strings"
)

var (
	hostport, scheme string
)

func newsQueryHandle(w http.ResponseWriter, r *http.Request) {
	hostport = r.Host + r.URL.Port()
	scheme = "http://"

	// get keyword
	keyword, err := parseKeyword(r)
	if err != nil {
		returnErr(err, w)
		return
	}
	fmt.Printf("[keyword] ==> \"%s\"\n", keyword)

	// crawl
	var cid, choice string
	if keyword == "" {
		nc, err := NewCrawlerNoKeyword()
		if err != nil {
			returnErr(err, w)
			return
		}
		cid, err = saveCard(nc.title, nc.subtitle, nc.coverURL(), nc.destURL())
		choice = nc.choice
		if err != nil {
			returnErr(err, w)
			return
		}
	} else {
		nc, err := NewCrawlerWithKeyword(keyword)
		if err != nil {
			returnErr(err, w)
			return
		}
		cid, err = saveCard(nc.title, nc.subtitle, nc.coverURL(), nc.destURL())
		if err != nil {
			returnErr(err, w)
			return
		}
	}

	// reply
	err = reply(keyword, cid, choice, w)
	if err != nil {
		returnErr(err, w)
		return
	}
}

func reply(keyword string, cid string, choice string, w http.ResponseWriter) error {
	resp := &Response{}
	replyText, err := getReply(keyword, choice)
	if err != nil {
		return err
	}
	resp.Slots = append(resp.Slots, Slots{"news_reply", replyText})
	resp.Slots = append(resp.Slots, Slots{"news_cid", cid})
	fmt.Printf("[机器人回复]%s \n", resp.ToBytes())
	w.Write(resp.ToBytes())
	return nil
}

func getReply(keyword string, choice string) (string, error) {
	links, err := readLinksFromJSON(GetRawNewsPath(keyword))
	if err != nil {
		return "", err
	}
	if choice != "" {
		links, err = chooseLinksFromKey(links, choice)
		if err != nil {
			return "", err
		}
	}
	if len(links) == 0 {
		return "", newsHandleError{"Too few links"}
	} else {
		reply := ""
		if keyword == "" {
			reply += "Latest news: \n"
		} else {
			reply += fmt.Sprintf("Latest news about %s: \n", keyword)
		}
		for i := range links {
			link := strings.ReplaceAll(links[i].Text, "#*", "")
			link = strings.ReplaceAll(link, "*#", "")
			reply += fmt.Sprintf("%d) %s\n", i+1, link)
		}
		return reply + "\nFor more details, click: ", nil
	}
}
