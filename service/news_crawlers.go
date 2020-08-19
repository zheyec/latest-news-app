package service

import (
	"fmt"
	"image"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/PuerkitoBio/goquery"
	"github.com/fogleman/gg"
)

const (
	newsURLNoKeyword   string        = "http://news.baidu.com/"
	//newsURLWithKeyword string        = "https://www.baidu.com/s?tn=news&rtt=4&bsst=1&cl=2&wd=%s&medium=1"
	newsURLWithKeyword string 		 = "https://www.baidu.com/s?ie=utf-8&cl=2&medium=1&rtt=4&bsst=1&rsv_dl=news_t_sk&tn=news&word=%s&rsv_sug3=4&rsv_sug4=170&rsv_sug1=4&rsv_sug2=0&inputT=1345"
	destURL            string        = "/news?month=%d&day=%d&nc=%s"
	destURLWithKeyword string        = "/news?month=%d&day=%d&keyword=%s"
	newsNum            int           = 10
	newsNumWithKeyword int           = 10
	newsPageLimit      int           = 2
	searchResDur       time.Duration = time.Hour * 1

	// FontPath - folder for fonts
	FontPath string = "static/fonts/"

	// TimgPath - folder for thunbnails
	TimgPath string = "data/thumbnails/"

	// RawNewsPath - folder for crawler results
	RawNewsPath string = "data/raw_news/"

	// StaticImagePath - folder for static images
	StaticImagePath string = "static/images/"

	// CardPath - folder for message cards
	CardPath string = "data/cards/"
)

// crawler when keyword is not provided
type crawlerNoKeyword struct {
	keyword  string
	title    string
	subtitle string
	month    string
	day      string
	timgURL  string
	links    []Link
	pageURL  string
	choice   string
}

// NewCrawlerNoKeyword - constructor
func NewCrawlerNoKeyword() (*crawlerNoKeyword, error) {
	var nc = new(crawlerNoKeyword)
	currentDate := strings.Fields(time.Now().Format("2006 01 02"))
	nc.month = currentDate[1]
	nc.day = currentDate[2]
	nc.title = "新闻头条"
	nc.subtitle = "看看最近发生的新闻头条"
	err := nc.initTimg()
	if err != nil {
		return nil, err
	}
	return nc, nc.initDest()
}

// initialize thumbnail
func (nc *crawlerNoKeyword) initTimg() error {
	oss_uploader, err := newOSSUploader()
	if err != nil {
		return err
	}

	// make a thumbnail
	path := GetTimgPath("", nc.month, nc.day)
	exist, _ := fileExist(path)
	if !exist {

		// write date
		bg, err := setBackground(GetStaticImagePath("timg_news.png"))
		if err != nil {
			return err
		}
		dc := gg.NewContextForImage(bg)
		err = dc.LoadFontFace(GetFontPath("SourceHanSansCN-Heavy.ttf"), 16)
		if err != nil {
			return err
		}
		dc.SetRGBA255(255, 255, 255, 127)
		dc.DrawString(fmt.Sprintf("%s/%s", nc.month, nc.day), 36, 88+16)
		dc.SavePNG(path)
	}

	// upload
	nc.timgURL, err = oss_uploader.upload_oss2("", nc.month, nc.day, path)
	return nil
}

// initialize destination page
func (nc *crawlerNoKeyword) initDest() error {
	newsNumTotal, err := readTotalFromJSON(GetRawNewsPath(""))
	if err != nil {
		return err
	}
	choice, err := randomChooseNums(newsNumTotal, newsNum)
	if err != nil {
		return err
	}
	month, err := strconv.Atoi(nc.month)
	if err != nil {
		return err
	}
	day, err := strconv.Atoi(nc.day)
	if err != nil {
		return err
	}
	nc.pageURL, nc.choice = scheme+hostport+fmt.Sprintf(destURL, month, day, choice), choice
	return nil
}

// return thumbnail URL
func (nc *crawlerNoKeyword) coverURL() string {
	return nc.timgURL
}

// return news page URL
func (nc *crawlerNoKeyword) destURL() string {
	return nc.pageURL
}

// BackgroundCrawler - crawls headlines in the background
type BackgroundCrawler struct {
}

// Start - start running in the background
func (bc *BackgroundCrawler) Start() {
	for {
		err := bc.crawl()
		fmt.Println("Background crawler executed")
		if err != nil {
			fmt.Println("Background crawler failed due to ", err.Error())
			return
		}
		ticker := time.NewTicker(time.Hour * 12)
		<-ticker.C
	}
}

func (bc *BackgroundCrawler) crawl() error {
	response, err := download(newsURLNoKeyword)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	result, err := bc.analyze(response)
	if err != nil {
		return err
	}

	err = writeLinks(GetRawNewsPath(""), result)
	if err != nil {
		return err
	}

	return nil
}

func (bc *BackgroundCrawler) analyze(resp *http.Response) ([]Link, error) {
	body, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return make([]Link, 0), err
	}

	// filter headlines using goquery
	var links []Link = make([]Link, 0)
	body.Find(".focuslistnews").Each(func(index int, item *goquery.Selection) {
		item.Find("a").Each(func(index int, item *goquery.Selection) {
			link, _ := item.Attr("href")
			linkText := item.Text()
			links = append(links, Link{linkText, link})
		})
	})
	return links, nil
}

// crawler when keyword is provided
type crawlerWithKeyword struct {
	keyword  string
	title    string
	subtitle string
	month    string
	day      string
	timgURL  string
	links    []Link
}

// NewCrawlerWithKeyword - constructor
func NewCrawlerWithKeyword(keyword string) (*crawlerWithKeyword, error) {
	var nc = new(crawlerWithKeyword)
	nc.keyword = keyword
	currentDate := strings.Fields(time.Now().Format("2006 01 02"))
	nc.month = currentDate[1]
	nc.day = currentDate[2]
	nc.title = fmt.Sprintf("Latest news about %s", keyword)
	nc.subtitle = fmt.Sprintf("Here is the latest news about %s", keyword)

	// check if crawling is necessary
	exist, _ := fileExist(GetRawNewsPath(keyword))
	if !exist {
		result, err := nc.crawlNews(keyword)
		if err != nil {
			return nc, err
		}
		err = writeLinks(GetRawNewsPath(keyword), result)
		if err != nil {
			return nc, err
		}

		// store results in local files
		fmt.Println("Created temp file: ", GetRawNewsPath(keyword))
		go removeWithDelay(GetRawNewsPath(keyword), searchResDur)
	}
	return nc, nc.initImg()
}

// initialize thumbnail
func (nc *crawlerWithKeyword) initImg() error {
	oss_uploader, err := newOSSUploader()
	if err != nil {
		return err
	}

	// make a thumbnail
	path := GetTimgPath(nc.keyword, nc.month, nc.day)
	exist, _ := fileExist(path)
	if !exist {

		// write date
		bg, err := setBackground(GetStaticImagePath("timg.png"))
		if err != nil {
			return err
		}
		dc := gg.NewContextForImage(bg)
		err = dc.LoadFontFace(GetFontPath("SourceHanSansCN-Heavy.ttf"), 16)
		if err != nil {
			return err
		}
		dc.SetRGBA255(255, 255, 255, 127)
		dc.DrawString(fmt.Sprintf("%s/%s", nc.month, nc.day), 36, 88+16)

		// write title
		err = dc.LoadFontFace(GetFontPath("SourceHanSansCN-Heavy.ttf"), 25)
		if err != nil {
			return err
		}
		dc.SetRGBA255(255, 255, 255, 255)
		lines := trim(nc.keyword)
		if len(lines) == 1 {
			dc.DrawStringAnchored(lines[0], 60, 50, 0.5, 0.5)
		} else {
			dc.DrawStringAnchored(lines[0], 60, 35, 0.5, 0.5)
			dc.DrawStringAnchored(lines[1], 60, 63, 0.5, 0.5)
		}
		dc.SavePNG(path)
		defer os.Remove(path)
	}

	// upload
	nc.timgURL, err = oss_uploader.upload_oss2(nc.keyword, nc.month, nc.day, path)
	return nil
}

// return thumbnail URL
func (nc *crawlerWithKeyword) coverURL() string {
	return nc.timgURL
}

// return news page URL
func (nc *crawlerWithKeyword) destURL() string {
	month, _ := strconv.Atoi(nc.month)
	day, _ := strconv.Atoi(nc.day)
	return scheme + hostport + fmt.Sprintf(destURLWithKeyword, month, day, url.QueryEscape(nc.keyword))
}

func (nc *crawlerWithKeyword) crawlNews(keyword string) ([]Link, error) {
	result := make([]Link, 0)
	currentURL := fmt.Sprintf(newsURLWithKeyword, url.QueryEscape(`“`+keyword+`”`))

	// get the first {newsPageLimit} pages
	for i := 0; i < newsPageLimit; i++ {
		response, err := download(currentURL)
		if err != nil {
			return result, err
		}
		defer response.Body.Close()
		links, nextURL, err := nc.analyze(response, i+1)
		if err != nil {
			return result, err
		}
		result = append(result, links...)

		// if links are not enough, go to next page
		if len(result) < newsNumWithKeyword {
			currentURL = nextURL
		} else {
			break
		}
	}
	return getFirstN(result, newsNumWithKeyword), nil
}

func (nc *crawlerWithKeyword) analyze(resp *http.Response, currentPN int) ([]Link, string, error) {
	body, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return make([]Link, 0), "", err
	}

	// select all search results using goquery
	var links []Link = make([]Link, 0)
	body.Find(".c-title").Each(func(index int, item *goquery.Selection) {
		item.Find("a").Each(func(index int, item *goquery.Selection) {
			plainText, _ := item.Html()
			linkText := strings.TrimSpace(strings.Split(plainText, "<nil>")[0])

			// results must include the keyword
			if strings.Index(linkText, "<em>") != -1 {
				replacedText := strings.ReplaceAll(linkText, "<em>", "#*")
				replacedText = strings.ReplaceAll(replacedText, "</em>", "*#")
				link, _ := item.Attr("href")
				links = append(links, Link{replacedText, link})
			}
		})
	})

	// find the URL to next page using goquery
	nextURL := ""
	body.Find("#page").Find("a").Each(func(index int, item *goquery.Selection) {
		if nextURL != "" {
			return
		}
		if index == currentPN {
			link, _ := item.Attr("href")
			nextURL = link
		}
	})

	return links, "https://www.baidu.com/" + nextURL, nil
}

func download(link string) (*http.Response, error) {
	fmt.Print("Trying to access link: ", link)
	client := &http.Client{}
	req, _ := http.NewRequest("GET", link, nil)

	// simulate browser headers
	headers := map[string]string{
		// 	"Pragma": "no-cache",
		// 	//"Accept-Encoding": "gzip, deflate, br",
		// 	"Accept-Language": "zh-CN,zh;q=0.9,en;q=0.8",
		// 	"Upgrade-Insecure-Requests": "1",
		"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.106 Safari/537.36",
		// 	"Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
		// 	"Cache-Control": "max-age=0",
		// 	"Connection": "keep-alive",
		// 	"Host": "www.baidu.com",
		// 	"Cookie": "BIDUPSID=EE160E3219E130A6A61D33A9C702B42E; PSTM=1590374703; BAIDUID=EE160E3219E130A6C05A997AD4832D80:FG=1; BD_UPN=123253; sugstore=0; BD_HOME=1; BDRCVFR[S4-dAuiWMmn]=I67x6TjHwwYf0; delPer=0; BD_CK_SAM=1; PSINO=3; BDORZ=B490B5EBF6F3CD402E515D22BCDA1598; COOKIE_SESSION=4_0_9_2_16_44_0_4_3_6_0_14_708113_0_2_0_1592463250_0_1592463248%7C9%23690411_10_1592274381%7C4; ZD_ENTRY=google; BDRCVFR[FhauBQh29_R]=mbxnW11j9Dfmh7GuZR8mvqV; BDRCVFR[C0p6oIjvx-c]=mbxnW11j9Dfmh7GuZR8mvqV; BDSVRTM=1269; H_PS_PSSID=",
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return resp, err
	}
	fmt.Printf(" (%s %s)\n", resp.Proto, resp.Status)

	return resp, nil
}

// GetRawNewsPath - folder for raw news json files
func GetRawNewsPath(keyword string) string {
	if keyword == "" {
		return path.Join(Cwd, RawNewsPath, "raw_news.json")
	}
	return path.Join(Cwd, RawNewsPath, fmt.Sprintf("temp_raw_news_%s.json", keyword))
}

func getTimgName(keyword string, month string, day string) string {
	if keyword == "" {
		return fmt.Sprintf("%s%s.png", month, day)
	} else {
		return fmt.Sprintf("temp_%s%s%s.png", keyword, month, day)
	}
}

// GetTimgPath - folder for thumbnails
func GetTimgPath(keyword string, month string, day string) string {
	return path.Join(Cwd, TimgPath, getTimgName(keyword, month, day))
}

// GetStaticImagePath - folder for static images
func GetStaticImagePath(filename string) string {
	return path.Join(Cwd, StaticImagePath, filename)
}

// GetFontPath - folder for fonts
func GetFontPath(fontname string) string {
	return path.Join(Cwd, FontPath, fontname)
}

// read an image, return type image.Image
func setBackground(path string) (image.Image, error) {
	srcFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	src, _, err := image.Decode(srcFile)
	if err != nil {
		return src, err
	}
	return src, nil
}

// trim down long keywords and split them into two lines, using ... if too long
func trim(text string) []string {
	var result []string
	length := 0
	oversize := false
	for _, r := range text {

		// every Chinese character counts as two English alphabets
		if unicode.Is(unicode.Scripts["Han"], r) {
			length += 2
		} else {
			length++
		}

		// the first line is up to 8 alphabets; the second line is up to 6
		if length <= 8 {
			if len(result) == 0 {
				result = append(result, "")
			}
			result[0] = result[0] + string(r)
		} else if length <= 14 {
			if len(result) == 1 {
				result = append(result, "")
			}
			result[1] = result[1] + string(r)
		} else {
			oversize = true
			break
		}
	}
	if oversize {
		result[1] += "…"
	}
	fmt.Println("[Trimmed to] ==> ", result)
	return result
}
