package main

import (
	"github.com/liuzl/ling"
)

var nlp = ling.MustNLP(ling.Norm)

func cut(line string) []string {
	d := ling.NewDocument(line)
	if err := nlp.Annotate(d); err != nil {
		return nil
	}
	return d.XTokens(ling.Norm)
}
