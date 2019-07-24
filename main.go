package main

import (
	"bufio"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/cheggaaa/pb"
	"github.com/liuzl/goutil"
	"github.com/liuzl/topk"
)

var (
	k = flag.Int("n", 10000, "k")
	i = flag.String("i", "input.txt.gz", "input file")
	o = flag.String("o", "output_%s.txt", "output file pattern")
)

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
	bar := pb.StartNew(count)
	for {
		line, e := br.ReadString('\n')
		if e == io.EOF {
			break
		}
		tokens := cut(strings.TrimSpace(line))
		for i := 1; i < len(tokens); i++ {
			if strings.TrimSpace(tokens[i]) == "" {
				continue
			}
			btk.InsertTokens(tokens[i:], 1)
		}
		bar.Increment()
	}
	bar.FinishPrint("done!")
	calc(btk, fmt.Sprintf(*o, "suffix"), count)
}

func calc(tk *topk.Stream, file string, total int) {
	var m = make(map[string]int)
	var m2 = make(map[string]map[string]int)
	for _, v := range tk.Keys() {
		m[v.Key] = v.Count
		terms := v.Items
		if len(terms) == 1 {
			continue
		}
		i := 1
		for ; strings.TrimSpace(terms[i]) == ""; i++ {
		}
		key := strings.TrimSpace(strings.Join(terms[i:], ""))
		prefix := terms[0]
		if m2[key] == nil {
			m2[key] = make(map[string]int)
		}
		m2[key][prefix] += v.Count
	}
	var records []*Record
	for _, v := range tk.Keys() {
		rec := &Record{Word: v.Key, Cnt: v.Count, Err: v.Error}
		rec.Flex = flex(m2, v)
		rec.Poly = poly(m, v)
		rec.Score = rec.Flex * rec.Poly
		records = append(records, rec)
	}
	out, err := os.Create(file)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	for _, r := range records {
		if r.Score < 1 || r.Cnt-r.Err < 10 {
			continue
		}
		fmt.Fprintf(out, "%s\t%d\t%d\t%f\t%f\t%f\n",
			r.Word, r.Cnt, r.Err, r.Poly, r.Flex, r.Score)
	}
}
