package language

import (
	"strconv"
	"strings"

	i18n "github.com/qor/i18n"
	"github.com/qor/i18n/backends/yaml"
)

var I18n *i18n.I18n

func init() {
	I18n = i18n.New(yaml.New("config/locale"))
}

type LangQ struct {
	Lang string
	Q    float64
}

func ParseAcceptLanguage(acptLang string) []LangQ {
	var lqs []LangQ

	langQStrs := strings.Split(acptLang, ",")
	for _, langQStr := range langQStrs {
		trimedLangQStr := strings.Trim(langQStr, " ")

		langQ := strings.Split(trimedLangQStr, ";")
		if len(langQ) == 1 {
			lq := LangQ{langQ[0], 1}
			lqs = append(lqs, lq)
		} else if len(langQ) > 1 {
			qp := strings.Split(langQ[1], "=")
			if len(qp) < 2 {
				lqs = append(lqs, LangQ{"en", 1})
			} else {
				q, err := strconv.ParseFloat(qp[1], 64)
				if err != nil {
					panic(err)
				}
				lq := LangQ{langQ[0], q}
				lqs = append(lqs, lq)
			}
		} else {
			lqs = append(lqs, LangQ{"en", 1})
		}
	}
	return lqs
}
