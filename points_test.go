package main

import (
	"bytes"
	"encoding/json"
	"testing"
	"text/tabwriter"

	"github.com/nordsieck/defect"
)

func TestClean(t *testing.T) {
	cases := map[string]string{
		"foo bar (123)":     "foo bar",
		"foo bar baz (543)": "foo bar baz",
	}

	for in, expected := range cases {
		defect.Equal(t, Clean(in), expected)
	}
}

func TestShrink(t *testing.T) {
	defect.DeepEqual(t, Shrink([]Small{
		{Name: "foo bar (123)", Wscid: 123},
		{Name: "foo bar baz (543)", Wscid: 543},
	}), []Small{
		{Name: "foo bar", Wscid: 123},
		{Name: "foo bar baz", Wscid: 543},
	})
}

var history = History{
	Type: "dancer",
	Dancer: Dancer{
		FirstName: "Amanda",
		LastName:  "Johnson",
		Wscid:     12345,
	},
	Placements: map[string][]Division{"West Coast Swing": {
		{
			Division:    DivisionName{Name: "Novice", Abbreviation: "NOV"},
			TotalPoints: 8,
			Competitions: []Result{
				{
					Role:   "leader",
					Points: 6,
					Event:  Event{Name: "Sea to Sky", Location: "Tacoma, WA", Date: "November 2016"},
					Result: "5",
				}, {
					Role:   "follower",
					Points: 1,
					Event:  Event{Name: "Swingtime in the Rockies", Location: "Denver, CO", Date: "July 2016"},
					Result: "F",
				}, {
					Role:   "follower",
					Points: 1,
					Event:  Event{Name: "Wild Wild Westie", Location: "Dallas, Texas", Date: "July 2016"},
					Result: "F",
				},
			},
		}, {
			Division:    DivisionName{Name: "Newcomer", Abbreviation: "NEW"},
			TotalPoints: 2,
			Competitions: []Result{{
				Role:   "follower",
				Points: 2,
				Event:  Event{Name: "The Challenge", Location: "Dallas, Texas", Date: "December 2015"},
				Result: "5",
			}},
		},
	}},
}

func TestHistory_Zero(t *testing.T) {
	var hist History
	defect.Equal(t, hist.Zero(), true)
	defect.Equal(t, history.Zero(), false)
}

func TestHistory_WriteTo(t *testing.T) {
	buf := &bytes.Buffer{}
	expected := `12345 Amanda Johnson

******************** West Coast Swing ********************

Novice: 8

Event                     Location       Date           Result  Points  Role
Sea to Sky                Tacoma, WA     November 2016  5       6       leader
Swingtime in the Rockies  Denver, CO     July 2016      F       1       follower
Wild Wild Westie          Dallas, Texas  July 2016      F       1       follower

Newcomer: 2

Event          Location       Date           Result  Points  Role
The Challenge  Dallas, Texas  December 2015  5       2       follower

`

	_, err := history.WriteTo(buf)
	defect.Equal(t, err, nil)
	defect.Equal(t, buf.String(), expected)

	buf.Reset()
	_, err = (&History{}).WriteTo(buf)
	defect.Equal(t, err, nil)
	defect.Equal(t, buf.String(), "No match found\n")
}

func TestUnMarshalHistory(t *testing.T) {
	raw := `
{
  "type":"dancer",
  "dancer":{
    "id":10000,
    "first_name":"Amanda",
    "last_name":"Johnson",
    "wscid":12345
  },
  "placements":{
    "West Coast Swing": [{
      "division": {"id":4,"name":"Novice","abbreviation":"NOV"},
      "total_points":8,
      "competitions":[{
        "role":"leader",
        "points":6,
        "event":{"id":79,"name":"Sea to Sky","location":"Tacoma, WA","url":"","date":"November 2016"},
        "result":"5"
      }, {
        "role":"follower",
        "points":1,
        "event":{"id":66,"name":"Swingtime in the Rockies","location":"Denver, CO","url":"","date":"July 2016"},
        "result":"F"
      }, {
        "role":"follower",
        "points":1,
        "event":{"id":225,"name":"Wild Wild Westie","location":"Dallas, Texas","url":"","date":"July 2016"},
        "result":"F"
    }]},{
      "division":{"id":3,"name":"Newcomer","abbreviation":"NEW"},
      "total_points":2,
      "competitions":[{
        "role":"follower",
        "points":2,
        "event":{"id":239,"name":"The Challenge","location":"Dallas, Texas","url":"","date":"December 2015"},
        "result":"5"
}]}]}}`

	buff := bytes.NewBufferString(raw)
	var hist History

	err := json.NewDecoder(buff).Decode(&hist)
	defect.Equal(t, err, nil)
	defect.DeepEqual(t, hist, history)
}

func TestHistory_WriteSummaryTo(t *testing.T) {
	expected := `12345 Amanda Johnson

******************** West Coast Swing ********************

Division  First          Last           Total  Role  1st  2nd  3rd  4th  5th  Final
Novice    July 2016      November 2016  8      l                         1/6  
                                               f                              2/2
Newcomer  December 2015  December 2015  2      f                         1/2  

`
	buf := &bytes.Buffer{}
	err := history.WriteSummaryTo(buf)
	defect.Equal(t, err, nil)
	defect.Equal(t, buf.String(), expected)
}

func TestDivision_Summary(t *testing.T) {
	type Case struct {
		d Division
		s *DivisionSummary
	}

	cases := []Case{{
		d: history.Placements[`West Coast Swing`][0],
		s: &DivisionSummary{
			Name:    `Novice`,
			From:    `July 2016`,
			To:      `November 2016`,
			Total:   8,
			Results: [2][6][2]int{{{}, {}, {}, {}, {1, 6}, {}}, {{}, {}, {}, {}, {}, {2, 2}}},
		},
	}, {
		d: history.Placements[`West Coast Swing`][1],
		s: &DivisionSummary{
			Name:    `Newcomer`,
			From:    `December 2015`,
			To:      `December 2015`,
			Total:   2,
			Results: [2][6][2]int{{}, {{}, {}, {}, {}, {1, 2}, {}}},
		},
	}}

	for _, c := range cases {
		defect.DeepEqual(t, c.d.Summary(), c.s)
	}
}

func TestDivisionSummary_WriteToTab(t *testing.T) {
	cases := map[string]*DivisionSummary{
		"Newcomer  December 2015  December 2015  2  f          1/2  \n": history.Placements[`West Coast Swing`][1].Summary(),
		"Novice  January 2015  June 2015  29  l  2/25  1/4        \n": &DivisionSummary{
			Name:    `Novice`,
			From:    `January 2015`,
			To:      `June 2015`,
			Total:   29,
			Results: [2][6][2]int{{{2, 25}, {1, 4}, {}, {}, {}, {}}, {}},
		},
		`Novice  July 2016  November 2016  8  l          1/6  
                                     f               2/2
`: history.Placements[`West Coast Swing`][0].Summary(),
	}

	for expected, ds := range cases {
		buf := &bytes.Buffer{}
		tw := tabwriter.NewWriter(buf, 0, 4, 2, ' ', 0)
		err := ds.WriteToTab(tw)
		tw.Flush()
		defect.Equal(t, err, nil)
		defect.Equal(t, buf.String(), expected)
	}

}
