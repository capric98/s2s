package s2s

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Translator interface {
	Trans(string) (string, error)
}

type youdao struct {
	appKey, appPass string
	from, to        string
	client          *http.Client
}

type ydResp struct {
	ErrCode     string   `json:"errorCode"`
	Translation []string `json:"translation"`
}

func NewYouDao(appKey, appPass string, from, to string) (Translator, error) {
	return &youdao{
		appKey:  appKey,
		appPass: appPass,
		from:    from,
		to:      to,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}, nil
}

func (t *youdao) Trans(text string) (r string, e error) {
	if text == "" {
		return
	}

	u, _ := uuid.NewUUID()
	currentT := strconv.FormatInt(time.Now().Unix(), 10)

	payload := url.Values{}
	payload.Set("q", text)
	payload.Set("from", t.from)
	payload.Set("to", t.to)
	payload.Set("appKey", t.appKey)
	payload.Set("salt", u.String())
	payload.Set("sign", t.genSign(text, u.String(), currentT))
	payload.Set("signType", "v3")
	payload.Set("curtime", currentT)

	req, _ := http.NewRequest("POST", "https://openapi.youdao.com/api", strings.NewReader(payload.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	var resp *http.Response
	resp, e = t.client.Do(req)
	if e != nil {
		return
	}
	defer resp.Body.Close()

	var results ydResp
	if e = json.NewDecoder(resp.Body).Decode(&results); e != nil {
		return
	}
	if results.ErrCode != "0" {
		//log.Println("salt:", u.String(), "curtime", currentT, "sign:", t.genSign(text, u.String(), currentT))
		e = errors.New("Translator: Remote server responses with error code: " + results.ErrCode)
		return
	} else {
		if len(results.Translation) == 0 {
			e = errors.New("Translator: Remote server responses with null translation.")
			return
		}
		r = results.Translation[0]
	}

	return
}

func (t *youdao) genSign(text, salt, curtime string) string {
	signStr := t.appKey + truncate(text) + salt + curtime + t.appPass
	//log.Println(signStr)
	checksum := sha256.Sum256([]byte(signStr))
	return hex.EncodeToString(checksum[:])
}

func truncate(text string) string {
	rt := []rune(text)
	l := len(rt)
	if l <= 20 {
		return text
	}

	return string(rt[:10]) + strconv.Itoa(l) + string(rt[l-10:l])
}
