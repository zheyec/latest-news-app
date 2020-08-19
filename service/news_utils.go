package service

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

// custom error type
type newsHandleError struct {
	Msg string
}

func (err newsHandleError) Error() string {
	return err.Msg
}

func returnErr(e error, w http.ResponseWriter) {
	fmt.Printf("Error: %s\n", e.Error())
	resp := &Response{}
	resp.ErrorNo = e.Error()
	fmt.Printf("[出错]%s \n", resp.ToBytes())
	w.Write(resp.ToBytes())
}

// Link - web links
type Link struct {
	Text string `json:"text"`
	Link string `json:"link"`
}

type linksWrapped struct {
	Total int    `json:"total"`
	Data  []Link `json:"data"`
}

// choose N numbers randomly and return base-64 encoded result
func randomChooseNums(total int, num int) (string, error) {
	if total < num {
		return "", newsHandleError{"Too few numbers"}
	}
	var choose []bool = make([]bool, total)
	for i := 0; i < num; i++ {
		j := rand.Int() % total
		for ; choose[j]; j = rand.Int() % total {
		}
		choose[j] = true
	}
	var result []int = make([]int, 0)
	for i := 0; i < total; i++ {
		if choose[i] {
			result = append(result, i)
		}
	}
	return encodeNums(result), nil
}

func chooseLinksFromKey(links []Link, encp string) ([]Link, error) {
	nums, err := decodeNums(encp)
	if err != nil {
		return nil, err
	}
	result := make([]Link, 0)
	for i := range nums {
		if nums[i] < 0 || nums[i] >= len(links) {
			return nil, newsHandleError{"Number for link out of range"}
		}
		result = append(result, links[nums[i]])
	}
	return result, nil
}

func encodeNums(nums []int) string {
	str := ""
	for i := range nums {
		if i != 0 {
			str += ","
		}
		str += strconv.Itoa(nums[i])
	}
	strBytes := []byte(str)
	return base64.URLEncoding.EncodeToString(strBytes)
}

func decodeNums(encp string) ([]int, error) {
	strBytes, err := base64.URLEncoding.DecodeString(encp)
	if err != nil {
		return nil, err
	}
	fields := strings.Split(string(strBytes), ",")
	nums := make([]int, 0)
	for i := range fields {
		num, err := strconv.Atoi(fields[i])
		if err != nil {
			return nums, err
		}
		nums = append(nums, num)
	}
	return nums, nil
}

// getFirstN - get first N links
func getFirstN(links []Link, num int) []Link {
	if len(links) < num {
		return links
	}
	return links[:num]
}

// writeLinks - write links to local files
func writeLinks(path string, links []Link) error {
	sampleText := ""
	if len(links) != 0 {
		sampleText = links[0].Text
	}
	fmt.Printf("Trying to write into.. %s with first link = [%s]\n", path, sampleText)
	data := linksWrapped{len(links), links}
	file, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(file)
	if err != nil {
		return err
	}
	return nil
}

// read links from json files
func readLinksFromJSON(path string) ([]Link, error) {
	data, err := unpackLinksJSON(path)
	if err != nil {
		return nil, err
	}
	return data.Data, nil
}

// read total number from json files
func readTotalFromJSON(path string) (int, error) {
	data, err := unpackLinksJSON(path)
	if err != nil {
		return -1, err
	}
	return data.Total, nil
}

func unpackLinksJSON(path string) (*linksWrapped, error) {
	data := linksWrapped{}
	rawNews, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(rawNews, &data)
	return &data, err
}

// Card - message card
type Card struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	CoverURL    string `json:"cover_url"`
	DestURL     string `json:"destination_url"`
}

// save message card to local files
func saveCard(title string, description string, coverURL string, destURL string) (string, error) {
	card := Card{title, description, coverURL, destURL}
	file, err := json.MarshalIndent(card, "", " ")
	if err != nil {
		return "", err
	}
	cid := getCardID(destURL)
	path := getCardPath(cid)
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	_, err = f.Write(file)
	if err != nil {
		return "", err
	}
	go removeWithDelay(path, time.Minute)
	return cid, nil
}

func getCard(cid string) (*Card, error) {
	card := Card{}
	cardBytes, err := ioutil.ReadFile(getCardPath(cid))
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(cardBytes, &card)
	return &card, err
}

func getCardPath(cid string) string {
	return path.Join(Cwd, CardPath, fmt.Sprintf("temp_card_%s.json", cid))
}

func getCardID(key string) string {
	hash := md5.Sum([]byte(key))
	return hex.EncodeToString(hash[:])
}

func removeWithDelay(path string, dur time.Duration) {
	timer := time.NewTimer(dur)
	<-timer.C
	fmt.Println("Removing temp file: ", path)
	os.Remove(path)
}

// RemoveAllTemp removes all files beginning with "temp_" in a specific folder
func RemoveAllTemp(path string) error {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}
	for _, f := range files {
		if !f.IsDir() && strings.HasPrefix(f.Name(), "temp_") {
			err := os.Remove(path + f.Name())
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// check if some file exists
func fileExist(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func parseKeyword(r *http.Request) (string, error) {
	rq := r.URL.RawQuery
	fmt.Println("Raw Query: ", rq)
	if !strings.HasPrefix(rq, "keyword=") && rq != "" {
		return "", &newsHandleError{"Bad URL Format"}
	}
	fields := strings.SplitN(rq, "=", 2)
	keyword := ""
	if len(fields) == 2 {
		keyword, _ = url.QueryUnescape(fields[1])
	}
	return keyword, nil
}
