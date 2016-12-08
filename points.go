package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
)

var (
	width = "5"

	matchToken = regexp.MustCompile(`name="_token" value="([^"]+)"`)

	search = flag.String("search", "", "Search for a dancer by name")
	name   = flag.String("name", "", "Get dancer's history by name")
	wsdcid = flag.Int("id", 0, "Get dancer's history by id")
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
		return
	} else if *name != "" || *wsdcid != 0 {
		token, err := GetToken()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}

		time.Sleep(time.Second)

		var query string
		if *name != "" && *wsdcid != 0 {
			query = fmt.Sprintf("%s (%d)", *name, *wsdcid)
		} else if *name != "" {
			query = *name
		} else if *wsdcid != 0 {
			query = strconv.Itoa(*wsdcid)
		}

		history, err := GetHistory(token, query)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}

		if _, err = history.WriteTo(os.Stdout); err != nil {
			fmt.Println(os.Stderr, err)
			return
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

func GetToken() (string, error) {
	resp, err := http.Get(`http://wsdc-points.us-west-2.elasticbeanstalk.com/lookup`)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	buff, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	matches := matchToken.FindSubmatch(buff)
	return string(matches[1]), nil
}

type History struct {
	Type       string                `json:"type"`
	Dancer     Dancer                `json:"dancer"`
	Placements map[string][]Division `json:"placements"`
}

// The number returned is garbage
func (h *History) WriteTo(w io.Writer) (int64, error) {
	if h.Zero() {
		_, err := fmt.Fprintln(w, "No match found")
		return 0, err
	}

	_, err := fmt.Fprintf(w, "%5d %s %s\n\n", h.Dancer.Wscid, h.Dancer.FirstName, h.Dancer.LastName)
	if err != nil {
		return 0, err
	}
	writer := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	for dance, divisions := range h.Placements {
		_, err = fmt.Fprintf(writer, "******************** %s ********************\n\n", dance)
		if err != nil {
			return 0, err
		}

		for _, division := range divisions {
			_, err = fmt.Fprintf(writer, "%s: %d\n\n", division.Division.Name, division.TotalPoints)
			if err != nil {
				return 0, err
			}

			_, err = fmt.Fprintf(writer, "Event\tLocation\tDate\tResult\tPoints\tRole\n")
			if err != nil {
				return 0, err
			}

			for _, result := range division.Competitions {
				_, err = fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%d\t%s\n", result.Event.Name, result.Event.Location,
					result.Event.Date, result.Result, result.Points, result.Role)
				if err != nil {
					return 0, err
				}
			}

			_, err = fmt.Fprintf(writer, "\n")
			if err != nil {
				return 0, err
			}
		}
	}
	err = writer.Flush()
	return 0, err
}

func (h *History) Zero() bool {
	if h.Type == "" && h.Dancer.FirstName == "" && h.Dancer.LastName == "" && h.Dancer.Wscid == 0 && len(h.Placements) == 0 {
		return true
	}
	return false
}

type Dancer struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Wscid     int    `json:"wscid"`
}

type Division struct {
	Division     DivisionName `json:"division"`
	TotalPoints  int          `json:"total_points"`
	Competitions []Result     `json:"competitions"`
}

type DivisionName struct {
	Name         string `json:"name"`
	Abbreviation string `json:"abbreviation"`
}

type Result struct {
	Role   string `json:"role"`
	Points int    `json:"points"`
	Event  Event  `json:"event"`
	Result string `json:"result"`
}

type Event struct {
	Name     string `json:"name"`
	Location string `json:"location"`
	Date     string `json:"date"`
}

func GetHistory(token, query string) (*History, error) {
	data := url.Values{
		"_token": {token},
		"q":      {query},
	}
	resp, err := http.PostForm(`http://wsdc-points.us-west-2.elasticbeanstalk.com/lookup/find`, data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var history History
	if err := json.NewDecoder(resp.Body).Decode(&history); err != nil {
		return nil, err
	}
	return &history, nil
}
