package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
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

func prepReq(method string, url string, referer string) *http.Request {
	req, _ := http.NewRequest(method, url, nil)
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/78.0.3904.70 Safari/537.36")
	req.Header.Add("Referer", referer)
	req.Header.Add("Origin", "https://www.bilibili.com")
	req.Header.Add("Sec-Fetch-Mode", "cors")
	req.Header.Add("Sec-Fetch-Site", "cross-site")

	return req
}

func download(url string, referer string, filename string, c *counter) {
	defer func() {
		if p := recover(); p != nil {
			fmt.Println("Download", url, "failed with:", p)
		}
	}()
	if url == "" {
		return
	}
	client := http.Client{}

	hresp, _ := client.Do(prepReq("HEAD", url, referer))
	length, _ := strconv.ParseInt(hresp.Header["Content-Length"][0], 10, 64)
	hresp.Body.Close()
	c.AddTask(length)

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

func main() {
	referer := os.Args[1]
	videoUrl := os.Args[2]
	audioUrl := os.Args[3]
	filename := os.Args[4]

	wg := sync.WaitGroup{}
	c := &counter{}

	wg.Add(2)
	go func() {
		download(videoUrl, referer, filename+".v", c)
		wg.Done()
	}()
	go func() {
		download(audioUrl, referer, filename+".a", c)
		wg.Done()
	}()

	for {
		c.mu.Lock()
		if c.total != 0 {
			c.mu.Unlock()
			break
		}
		c.mu.Unlock()
		time.Sleep(time.Second)
	}

	go func() {
		for {
			c.mu.Lock()
			tmp := float64(c.done)
			speed := float64(c.done - c.last)
			c.last = c.done
			if c.done == c.total {
				return
			}
			c.mu.Unlock()
			//fmt.Println(tmp, c.total)
			fmt.Printf("Finish: %5.2fM/%5.2fM  %5.2f M/s\r", tmp/1024/1024, float64(c.total/1024/1024), speed/1024/1024)
			time.Sleep(time.Second)
		}
	}()
	wg.Wait()
	fmt.Println()
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
