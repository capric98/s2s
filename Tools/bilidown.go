package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	CHUNK  = 128000
	THREAD = 8
)

type counter struct {
	done  int64
	total int64
	last  int64
	mu    sync.Mutex
}

type BiliAPI struct {
	Code   int          `json:"code"`
	Msg    string       `json:"message"`
	Result DashResult   `json:"result"`
	Data   LegacyResult `json:"data"`
}

type LegacyAPI struct {
	Data []LegacyResult `json:"data"`
}

type DashResult struct {
	Type string `json:"type"`
	DUrl []DUrl `json:"durl"`
}
type LegacyResult struct {
	Cid    int    `json:"cid"`
	Title  string `json:"title"`
	Format string `json:"format"`
	DUrl   []DUrl `json:"durl"`
}
type DUrl struct {
	Size int64  `json:"size"`
	Url  string `json:"url"`
}

var (
	ck     = flag.String("c", "cookie.txt", "cookie file")
	client = http.Client{}
)

func main() {
	flag.Parse()

	cookie, _ := ioutil.ReadFile(*ck)
	sc := string(cookie)
	for sc[len(sc)-1] == '\n' {
		sc = sc[:len(sc)-1]
	}

	for _, v := range flag.Args() {
		download(sc, v)
	}

}

func getCID(aid string, part int) string {
	creq, _ := http.NewRequest("GET", "https://api.bilibili.com/x/player/pagelist?aid="+aid+"&jsonp=jsonp", nil)
	cresp, err := client.Do(creq)
	if err != nil {
		log.Println(err)
		return ""
	}
	var cidjson LegacyAPI
	err = json.NewDecoder(cresp.Body).Decode(&cidjson)
	cresp.Body.Close()
	if err != nil {
		log.Println(err)
		return ""
	}
	return strconv.Itoa(cidjson.Data[part].Cid)
}

func getTitle(aid string, cid string) string {
	var titlejson BiliAPI
	req, _ := http.NewRequest("GET", "https://api.bilibili.com/x/web-interface/view?aid="+aid+"&cid="+cid, nil)
	tresp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return ""
	}
	err = json.NewDecoder(tresp.Body).Decode(&titlejson)
	tresp.Body.Close()
	if err != nil {
		log.Println(err)
		return ""
	}
	return NameRegularize(titlejson.Data.Title)
}

func download(cookie string, id string) {
	c := &counter{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var url, referer, filename string
	var part int
	var aresp BiliAPI
	var req *http.Request

	switch id[:2] {
	case "av", "AV", "Av", "aV":
		referer = "https://www.bilibili.com/video/" + id

		aid := id[2:]
		if strings.Contains(aid, "?") {
			aid = aid[:strings.Index(aid, "?")]
			part, _ = strconv.Atoi(id[strings.Index(id, "?")+3:])
		}

		if part != 0 {
			part = part - 1
		}
		cid := getCID(aid, part)

		filename = getTitle(aid, cid)
		if part != 0 {
			filename += "p" + strconv.Itoa(part)
		}

		req, _ = http.NewRequest("GET", "https://api.bilibili.com/x/player/playurl?avid="+aid+"&cid="+cid+"&qn=116&otype=json", nil)
	case "ep", "EP", "Ep", "eP":
		referer = "https://www.bilibili.com/bangumi/play/" + id

		epid := id[2:]
		pageReq, _ := http.NewRequest("GET", referer, nil)
		pageResp, err := client.Do(pageReq)
		if err != nil {
			log.Println(err)
			return
		}
		pagebyte, err := ioutil.ReadAll(pageResp.Body)
		pageResp.Body.Close()
		if err != nil {
			log.Println(err)
			return
		}
		page := string(pagebyte)
		page = page[strings.Index(page, `class="av-link">`)+16:]
		aid := page[2:strings.Index(page, `<`)]

		cid := getCID(aid, 0)
		filename = getTitle(aid, cid)

		req, _ = http.NewRequest("GET", "https://api.bilibili.com/pgc/player/web/playurl?cid="+cid+"&qn=116&otype=json&avid="+aid+"&ep_id="+epid, nil)
	default:
		log.Println(id, "is not a valid id.")
		return
	}

	req.Header.Add("Referer", referer)
	req.Header.Add("Cookie", cookie)
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/78.0.3904.70 Safari/537.36")
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()
	if e := json.NewDecoder(resp.Body).Decode(&aresp); e != nil {
		log.Println(e)
		return
	}

	var durls []DUrl
	if aresp.Code != 0 {
		log.Println("API response:", aresp.Msg)
		return
	}
	if aresp.Result.DUrl != nil {
		durls = aresp.Result.DUrl
		filename += "." + toFormat(aresp.Result.Type)
	} else {
		durls = aresp.Data.DUrl
		filename += "." + toFormat(aresp.Data.Format)
	}

	var maxDurl DUrl
	var max int64
	for _, v := range durls {
		//log.Println(v.Size)
		if v.Size > max {
			max = v.Size
			maxDurl = v
		}
	}
	url = maxDurl.Url
	c.AddTask(max)
	log.Println("Downloading:", filename)

	go func() {
		t := time.NewTicker(time.Second)
		defer t.Stop()
		defer fmt.Print("\n")
		for {
			select {
			case <-t.C:
				c.mu.Lock()
				fmt.Print("\rFinish:", addUnit(c.done), "/", addUnit(c.total), " Speed:", addUnit(c.done-c.last), "/s")
				c.last = c.done
				c.mu.Unlock()
			case <-ctx.Done():
				return
			}
		}
	}()

	multiThreadDown(url, referer, filename, c)
}

func prepReq(method string, url string, referer string) *http.Request {
	req, _ := http.NewRequest(method, url, nil)
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/78.0.3904.70 Safari/537.36")
	req.Header.Add("Referer", referer)
	req.Header.Add("Origin", "https://www.bilibili.com")
	req.Header.Add("Sec-Fetch-Mode", "cors")
	req.Header.Add("Sec-Fetch-Site", "cross-site")

	return req
}

func multiThreadDown(url string, referer string, filename string, c *counter) {
	defer func() {
		if p := recover(); p != nil {
			fmt.Println("Download", url, "failed with:", p)
		}
	}()
	if url == "" {
		return
	}

	hresp, _ := client.Do(prepReq("HEAD", url, referer))
	length, _ := strconv.ParseInt(hresp.Header["Content-Length"][0], 10, 64)
	hresp.Body.Close()
	//c.AddTask(length)

	f, _ := os.Create(filename)
	defer f.Close()
	it := make(chan int64)
	go func() {
		defer close(it)
		for i := int64(0); i < length; i += CHUNK {
			it <- i
		}
	}()

	wg := sync.WaitGroup{}
	for i := 0; i < THREAD; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			//client := http.Client{}
			for pos := range it {
				end := pos + CHUNK
				if end >= length {
					end = length + 1
				}
				req := prepReq("GET", url, referer)
				req.Header.Add("Range", "bytes="+strconv.FormatInt(pos, 10)+"-"+strconv.FormatInt(end-1, 10))
				resp, _ := client.Do(req)
				body, _ := ioutil.ReadAll(resp.Body)
				resp.Body.Close()
				n, _ := f.WriteAt(body, pos)
				c.Finish(int64(n))
			}
		}()
	}
	wg.Wait()
}

func (c *counter) AddTask(l int64) {
	c.mu.Lock()
	c.total += l
	c.mu.Unlock()
}

func (c *counter) Finish(l int64) {
	c.mu.Lock()
	c.done += l
	c.mu.Unlock()
}

func addUnit(n int64) string {
	count := 0
	fn := float64(n)
	for fn >= 1024 {
		fn /= 1024
		count += 1
	}
	return fmt.Sprintf("%7.1f", fn) + unit(count)
}

func unit(n int) string {
	switch n {
	case 0:
		return "B"
	case 1:
		return "KiB"
	case 2:
		return "MiB"
	case 3:
		return "GiB"
	case 4:
		return "TiB"
	default:
		return "???"
	}
}

func NameRegularize(name string) string {
	name = strings.ReplaceAll(name, ":", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "*", "_")
	name = strings.ReplaceAll(name, "?", "_")
	name = strings.ReplaceAll(name, "\"", "_")
	name = strings.ReplaceAll(name, "<", "_")
	name = strings.ReplaceAll(name, ">", "_")
	name = strings.ReplaceAll(name, "|", "_")
	name = strings.ReplaceAll(name, "\n", "_")
	name = strings.ReplaceAll(name, "\r", "_")
	name = strings.ReplaceAll(name, " ", "_")
	if len(name) > 255 {
		name = name[:255]
	}
	return name
}

func toFormat(s string) string {
	switch {
	case strings.Contains(s, "flv"), strings.Contains(s, "FLV"), strings.Contains(s, "Flv"):
		return "flv"
	// case strings.Contains(s, "mp4"), strings.Contains(s, "MP4"), strings.Contains(s, "m4a"), strings.Contains(s, "M4a"), strings.Contains(s, "M4A"):
	// 	return "mp4"
	default:
		return "mp4"
	}
}
