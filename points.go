package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var (
	width = "5"

	search = flag.String("search", "", "Search for a dancer by name")
)

type Small struct {
	Name  string `json:"name"`
	Wscid int    `json:"wscid"`
}

func (s Small) Print() { fmt.Printf("%"+width+"v: %v\n", s.Wscid, s.Name) }

func main() {
	flag.Parse()

	if *search != "" {
		resp, err := http.Get(`http://wsdc-points.us-west-2.elasticbeanstalk.com/lookup/autocomplete?q="` + url.QueryEscape(*search) + `"`)
		defer resp.Body.Close()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}

		var results []Small
		if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}

		results = Shrink(results)
		for _, r := range results {
			r.Print()
		}
	}
}

func Clean(s string) string { return s[:strings.LastIndex(s, " ")] }
func Shrink(ss []Small) []Small {
	for i := range ss {
		ss[i].Name = Clean(ss[i].Name)
	}
	return ss
}
