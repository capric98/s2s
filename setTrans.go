package s2s

import "log"

func SetTrans(r [][]Sentence, from, to string, opt ...interface{}) {
	tr, e := NewYouDao(opt[0].(string), opt[1].(string), from, to)
	if e != nil {
		log.Println("Failed to init Translator:", e)
		return
	}

	// tts, _ := tr.Trans("天気が良いから、散歩しましょう。")
	// log.Println("Test Trans:", tts)

	for _, speaker := range r {
		for k := range speaker {
			sentence := speaker[k]
			speaker[k].Trans = func() string {
				for i := 0; i < 3; i++ {
					if ts, e := tr.Trans(sentence.Content()); e == nil {
						return ts
					} else {
						log.Println("On Trans", sentence.Content(), ":", e)
					}
				}
				return ""
			}()
			log.Println(sentence.Content(), "->", speaker[k].Trans)
		}
	}
}
