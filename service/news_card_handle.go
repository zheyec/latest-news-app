package service

import (
	"fmt"
	"net/http"
)

func newsCardHandle(w http.ResponseWriter, r *http.Request) {
	// get card
	cid := r.URL.Query().Get("cid")
	card, err := getCard(cid)
	if err != nil {
		returnErr(err, w)
		return
	}

	// return card
	writeCard(card, "dummy", w)
}

func writeCard(card *Card, dummy string, w http.ResponseWriter) {
	resp := &Response{}
	cardWrapped := map[string]utf8String{
		"title":           utf8String(card.Title),
		"description":     utf8String(card.Description),
		"cover_url":       utf8String(card.CoverURL),
		"destination_url": utf8String(card.DestURL),
	}
	resp.Messages = append(resp.Messages, ResponseMsg{cardWrapped, "share_link"})
	resp.Slots = append(resp.Slots, Slots{"news_dummy", dummy})
	fmt.Printf("[Response]%s \n", resp.ToBytes())
	w.Write(resp.ToBytes())
}
