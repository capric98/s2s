package s2s

import (
	"io"
	"os"
	"strconv"
	"time"
)

type Style struct {
	Name    string
	Content string
}

type Dialogue struct {
	Layer                     int
	Start, End                time.Duration
	Style, Name               string
	MarginL, MarginR, MarginV int
	Effect, Text              string
}

var (
	slen   int
	styles []Style

	minorShift   = 300 * time.Millisecond
	maxLineWidth = 25
)

func initStyle() {
	styles = []Style{
		Style{
			Name:    "Default",
			Content: "微软雅黑,16,&H00FFFFFF,&H0300FFFF,&H00000000,&H02000000,0,0,0,0,100,100,0,0,3,1,1,1,10,10,10,1\n",
		},
	}
	slen = len(styles)
}

func OutputSubtitle(result [][]Sentence, out string, verbose, withTrans bool) (e error) {
	initStyle()

	fo, e := os.Create(out)
	if e != nil {
		return
	}
	defer fo.Close()

	// Write Header.
	_, _ = fo.Write([]byte(scriptInfo))
	// Write Format/Style
	for _, v := range styles {
		bswrite(fo, "Style: ", v.Name, ",", v.Content)
	}
	_, _ = fo.Write([]byte(events))

	var dia Dialogue
	var w Word
	for k := range result {
		speaker := result[k]
		dia.Style = styles[k%slen].Name
		for _, sentence := range speaker {
			if sentence.IsEnd() {
				// In case of null sentence.
				continue
			}

			w = sentence.Pop()
			if dia.End < w.End-minorShift {
				dia.End = w.End - minorShift
			}
			queue := ""
			for !sentence.IsEnd() {
				queue += w.C
				if verbose {
					dia.Start = dia.End
					dia.End = w.End
					dia.Text = prepText(queue, sentence.Trans, withTrans)
					dia.writeTo(fo)
				}
				w = sentence.Pop()
			}
			queue += w.C
			dia.Start = dia.End
			dia.End = w.End + minorShift
			dia.Text = prepText(queue, sentence.Trans, withTrans)
			dia.writeTo(fo)

			sentence.Reset()
		}
	}

	return
}

func prepText(text, trans string, withTrans bool) string {
	result := []rune{}
	rt := []rune(text)
	for len(rt) > maxLineWidth {
		result = append(result, rt[:maxLineWidth]...)
		result = append(result, []rune("\\N")...)
		rt = rt[maxLineWidth:]
	}
	result = append(result, rt...)
	if withTrans {
		result = append(result, []rune("\\N")...)
		rt = []rune(trans)
		for len(rt) > maxLineWidth {
			result = append(result, rt[:maxLineWidth]...)
			result = append(result, []rune("\\N")...)
			rt = rt[maxLineWidth:]
		}
		result = append(result, rt...)
	}
	return string(result)
}

func (d Dialogue) writeTo(w io.Writer) {
	if d.End <= d.Start {
		return
	}
	bswrite(w, "Dialogue: ")
	bswrite(w, strconv.Itoa(d.Layer))
	bswrite(w, ",", toSubTime(d.Start), ",", toSubTime(d.End))
	bswrite(w, ",", d.Style, ",", d.Name)
	bswrite(w, ",", strconv.Itoa(d.MarginL), ",", strconv.Itoa(d.MarginR), ",", strconv.Itoa(d.MarginV))
	bswrite(w, ",", d.Effect)
	bswrite(w, ",", d.Text, "\n")
}

func toSubTime(t time.Duration) (s string) {
	sec := int(t.Seconds())
	ms := (t - time.Duration(sec)*time.Second).Milliseconds()
	s += strconv.Itoa(sec/3600) + ":"
	sec %= 3600
	if sec < 60 {
		s += "0"
	}
	s += strconv.Itoa(sec/60) + ":"
	sec %= 60
	if sec < 10 {
		s += "0"
	}
	s += strconv.Itoa(sec) + "."
	if ms < 100 {
		s += "0"
	}
	s += strconv.Itoa(int(ms) / 10)
	return
}

func bswrite(w io.Writer, s ...string) {
	for k := range s {
		_, _ = w.Write([]byte(s[k]))
	}

}
