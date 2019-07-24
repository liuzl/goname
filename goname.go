package main

import (
	"math"
	"strings"

	"github.com/liuzl/topk"
)

type Record struct {
	Word  string
	Cnt   int
	Err   int
	Poly  float64
	Flex  float64
	Score float64
}

func Entropy(p []float64) float64 {
	if len(p) == 0 {
		return 1 //math.Maxfloat64
	}
	var e float64
	for _, v := range p {
		if v != 0 { // Entropy needs 0 * log(0) == 0
			e -= v * math.Log(v)
		}
	}
	return e
}

func poly(m map[string]int, v topk.Element) float64 {
	if len(v.Items) == 1 {
		return 1.0
	}
	i := 1
	for ; strings.TrimSpace(v.Items[i]) == ""; i++ {
	}
	short := m[strings.TrimSpace(strings.Join(v.Items[i:], ""))]
	if short == 0 {
		return 1.0
	}
	return float64(m[v.Key]) / float64(short)
}

func flex(m2 map[string]map[string]int, v topk.Element) float64 {
	if m2[v.Key] == nil {
		return 1.0
	}
	return entropy(m2[v.Key])
}

func entropy(m map[string]int) float64 {
	var p []float64
	var total float64
	for _, v := range m {
		total += float64(v)
	}
	for _, v := range m {
		p = append(p, float64(v)/total)
	}
	return Entropy(p)
}
