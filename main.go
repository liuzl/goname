package main

import (
	"bufio"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/cheggaaa/pb"
	"github.com/dgryski/go-topk"
	"github.com/liuzl/goutil"
	"github.com/liuzl/ling"
)

var (
	k = flag.Int("n", 10000, "k")
	i = flag.String("i", "input.txt.gz", "input file")
	c = flag.Bool("c", false, "use count flag")
	o = flag.String("o", "output_%s.txt", "output file pattern")
)

var nlp = ling.MustNLP(ling.Norm)

func main() {
	flag.Parse()
	count, err := goutil.FileLineCount(*i)
	if err != nil {
		log.Fatal(err)
	}
	f, err := os.Open(*i)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	var br *bufio.Reader
	if strings.HasSuffix(strings.ToLower(*i), ".gz") {
		if gr, err := gzip.NewReader(f); err != nil {
			log.Fatal(err)
		} else {
			br = bufio.NewReader(gr)
		}
	} else {
		br = bufio.NewReader(f)
	}
	btk := topk.New(*k)
	ftk := topk.New(*k)
	bar := pb.StartNew(count)
	for {
		line, e := br.ReadString('\n')
		if e == io.EOF {
			break
		}
		line = strings.TrimSpace(line)
		items := strings.Fields(line)
		line = items[0]
		cnt := 1
		if *c {
			if len(items) > 1 {
				if n, err := strconv.Atoi(items[1]); err == nil {
					cnt = n
				}
			}
		}
		d := ling.NewDocument(line)
		if err := nlp.Annotate(d); err != nil {
			log.Printf("%s, %+v\n", line, err)
			continue
		}
		tokens := d.XRealTokens(ling.Norm)
		for i := 1; i < len(tokens); i++ {
			btk.Insert(strings.Join(tokens[i:], " "), cnt)
			ftk.Insert(strings.Join(tokens[:i], " "), cnt)
		}
		bar.Increment()
	}
	bar.FinishPrint("done!")
	output(ftk, fmt.Sprintf(*o, "prefix"))
	output(btk, fmt.Sprintf(*o, "suffix"))
	calc(btk)
}

func output(tk *topk.Stream, file string) {
	out, err := os.Create(file)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	for _, v := range tk.Keys() {
		fmt.Fprintln(out, v.Key, v.Count, v.Error)
	}
}

type Record struct {
	Word  string
	Cnt   int
	Err   int
	Poly  float64
	Flex  float64
	Score float64
}

func poly(m map[string]int, w string) float64 {
	terms := strings.Fields(w)

	if len(terms) == 1 {
		return 1.0
	}
	short := m[strings.Join(terms[1:], " ")]
	if short == 0 {
		return 1.0
	}
	return float64(m[w]) / float64(short)
}

func flex(m2 map[string]map[string]int, w string) float64 {
	if m2[w] == nil {
		return 1.0
	}
	return entropy(m2[w])
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

func calc(tk *topk.Stream) {
	var m = make(map[string]int)
	var m2 = make(map[string]map[string]int)
	for _, v := range tk.Keys() {
		terms := strings.Fields(v.Key)
		m[v.Key] = v.Count
		if len(terms) == 1 {
			continue
		}
		key := strings.Join(terms[1:], " ")
		prefix := terms[0]
		if m2[key] == nil {
			m2[key] = make(map[string]int)
		}
		m2[key][prefix] += v.Count
	}
	var records []*Record
	for _, v := range tk.Keys() {
		rec := &Record{Word: v.Key, Cnt: v.Count, Err: v.Error}
		rec.Flex = flex(m2, v.Key)
		rec.Poly = poly(m, v.Key)
		rec.Score = rec.Flex * rec.Poly
		records = append(records, rec)
	}
	sort.Slice(records, func(i, j int) bool {
		return records[i].Score > records[j].Score
	})
	out, err := os.Create("village.out")
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	for _, r := range records {
		if r.Score < 1 {
			continue
		}
		fmt.Fprintf(out, "%s\t%d\t%d\t%f\t%f\t%f\n", r.Word, r.Cnt, r.Err, r.Poly, r.Flex, r.Score)
	}
}
