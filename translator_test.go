package s2s

import (
	"fmt"
	"os"
	"testing"
)

var (
	longText  = "聴解試験を始める前に、音を聞いてください。また、問題用紙を開けないでください。音がよく聞こえない時は、手を挙げてください。"
	longTran  = "在开始听力考试之前，请听声音。另外，请不要打开试卷。听不清声音的时候，请举手。"
	shortText = "天気が良いから、散歩しましょう。"
	shortTran = "天气很好，去散步吧。"
)

func newyd() (Translator, error) {
	appKey := os.Getenv("APPKEY")
	appPass := os.Getenv("APPPASS")
	return NewYouDao(appKey, appPass, "ja", "zh-CHS")
}

func TestLongText(t *testing.T) {
	translator, e := newyd()
	if e != nil {
		fmt.Println(e)
		t.Fail()
	}
	tran, e := translator.Trans(longText)
	if e != nil {
		fmt.Println(e)
		t.Fail()
	}
	if tran != longTran {
		fmt.Println("Resp:", tran, "rather than:", longTran)
		t.Fail()
	}
}
func TestShortText(t *testing.T) {
	translator, e := newyd()
	if e != nil {
		fmt.Println(e)
		t.Fail()
	}
	tran, e := translator.Trans(shortText)
	if e != nil {
		fmt.Println(e)
		t.Fail()
	}
	if tran != shortTran {
		fmt.Println("Resp:", tran, "rather than:", shortTran)
		t.Fail()
	}
}
