// AtB route planner cli in go.
// Date: 12.04.2019
// Author: Jørgen Bele Reinfjell
// Description: AtB route plannner cli to get next departure

// TODO: Support other transport types; boats, ...

package atb

import (
	"fmt"
	"github.com/anaskhan96/soup"
	"github.com/levigross/grequests"
	"strconv"
	"strings"
	"time"
)

// URL for the 'suggestions' endpoint.
const SuggestionsURL = "https://rp.atb.no/scripts/TravelMagic/TravelMagicWE.dll/StageJSON"

type SuggestionRes struct {
	Query       string   `json:"query"`
	Suggestions []string `json:"suggestions"`
}

func GetSuggestions(query string) ([]string, error) {
	ro := &grequests.RequestOptions{
		Params: map[string]string{"query": query},
	}

	resp, err := grequests.Get(SuggestionsURL, ro)
	if err != nil {
		return nil, err
	}

	var sr SuggestionRes
	err = resp.JSON(&sr)
	return sr.Suggestions, err
}

// URL for the 'departures' endpoint.
const DeparturesURL = `https://rp.atb.no/scripts/TravelMagic/TravelMagicWE.dll/svar`

type TransportType int

const (
	TransportBus TransportType = iota
	TransportWalking
)

type Transport struct {
	Type     TransportType
	WalkText string
	LineNum  int
	Start    time.Time
}

type Departure struct {
	Start    time.Time
	End      time.Time
	Changes  int
	Fare     string
	Duration time.Duration
	Route    []Transport
}

func getDeparturesResp(dir int, from, to, dtime, ddate string) (resp string, err error) {
	ro := &grequests.RequestOptions{
		UserAgent: "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/73.0.3683.103 Safari/537.36",
		Params: map[string]string{
			"direction": "1", //string(dir),
			"from":      from,
			"to":        to,
			"time":      dtime,
			"date":      ddate,
			"search":    "Show travel suggestions",
			"lang":      "en",
			//"through":       "",
			//"throughpause":  "",
			//"changepenalty": "1",
			//"changepause":   "0",
			//"linjer":        "",
			//"destinations":  "",
		},
	}

	gresp, err := grequests.Get(DeparturesURL, ro)
	if err != nil {
		return
	}
	resp = gresp.String()
	return
}

func dateTimeMerge(datestr, timestr string) (time.Time, error) {
	// Use the date of departure together with the start/end time
	// to convert to a time.Time object.
	timeLayout := "Monday 2 January 2006 15:04"
	t, err := time.Parse(timeLayout, datestr+timestr)
	return t, err
}

func GetDepartures(dir int, from, to, dtime, ddate string) (deps []Departure, err error) {
	defer func() {
		if r := recover(); r != nil {
			// Error occoured, error that caused is stored in 'err'
			// and will be returned.
		}
	}()

	// Closures to ease error handling.
	find := func(root soup.Root, args ...string) (res soup.Root) {
		// Panic if err != nil, will be catched by deferred recover() as
		// seen above.
		if err != nil {
			panic(err)
			return
		}
		res = root.Find(args...)
		err = res.Error
		return
	}

	findStrict := func(root soup.Root, args ...string) (res soup.Root) {
		if err != nil {
			panic(err)
			return
		}
		res = root.FindStrict(args...)
		err = res.Error
		return
	}

	findAll := func(root soup.Root, args ...string) (res []soup.Root) {
		if err != nil {
			panic(err)
			return
		}
		res = root.FindAll(args...)
		for _, r := range res {
			if r.Error != nil {
				err = r.Error
				panic(err)
				break
			}
		}
		return
	}

	findAllStrict := func(root soup.Root, args ...string) (res []soup.Root) {
		if err != nil {
			panic(err)
			return
		}
		res = root.FindAllStrict(args...)
		for _, r := range res {
			if r.Error != nil {
				err = r.Error
				panic(err)
				break
			}
		}
		return
	}

	html, err := getDeparturesResp(dir, from, to, dtime, ddate)
	if err != nil {
		return
	}

	doc := soup.HTMLParse(html)
	if err = doc.Error; err != nil {
		return
	}

	mainContent := findStrict(doc, "div", "class", "maincontent")

	// Used when parsing start and end times.
	date := findStrict(doc, "h2", "class", "tm-alpha tm-reiseforslag-header")
	if err = date.Error; err != nil {
		return
	}

	dateStr := date.Text()
	resultWrappers := findAllStrict(mainContent, "div", "class", "tm-result-wrapper")

	for _, rw := range resultWrappers {
		var d Departure

		// The tm-block-b span contains:
		//     start and end time, duration, changes and fare
		blockB := findStrict(rw, "span", "class", "tm-block-b")

		// tm-block-b contains two tm-result-time-wrapper elements
		// where the first one is the start time, and the second
		// one is the end time.
		unpackTime := func(res []soup.Root) (start, end time.Time, err error) {
			// Use the date of departure together with the start/end time
			// to convert to a time.Time object.
			start, err = dateTimeMerge(dateStr, res[0].Text())
			if err != nil {
				return
			}
			end, err = dateTimeMerge(dateStr, res[1].Text())
			return
		}
		d.Start, d.End, err = unpackTime(blockB.FindAll("span", "class", "tm-result-fratil"))
		if err != nil {
			return
		}

		// tm-result-details-extra contains three tm-result-info elements:
		// 1. duration, 2. changes, 3. fare
		// The 'tm-result-info-val' element contains the string describing each
		// element previously described.
		extraDetails := findStrict(blockB, "span", "class", "tm-inline-block tm-result-details-extra")

		spanClasses := []string{"tm-result-value-time", "tm-result-value-change", "tm-result-value-price"}
		values := [3]soup.Root{}
		for i, className := range spanClasses {
			values[i] = find(extraDetails, "span", "class", className).Find("span", "class", "tm-result-info-val")
		}

		parseDuration := func(duration string) (d time.Duration, err error) {
			durationSplice := strings.Split(duration, ":")
			hours, err := strconv.Atoi(durationSplice[0])
			if err != nil {
				return
			}
			minutes, err := strconv.Atoi(durationSplice[1]) // err is returned
			d = time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute
			return
		}

		unpackExtraDetails := func(res [3]soup.Root) (duration time.Duration, changes int, fare string, err error) {
			duration, err = parseDuration(res[0].Text())
			if err != nil {
				return
			}
			changes, err = strconv.Atoi(res[1].Text()) // err is returned
			if err != nil {
				// Edge case: when the trip consists of walking only (how I discovered
				// this) or if there is no changes, then a '-'  might be displayed
				// instead. This fixes this issue, but is not the best solution.

				// count as 0
				changes = 0
				err = nil
			}

			fare = res[2].Text()
			return
		}

		d.Duration, d.Changes, d.Fare, err = unpackExtraDetails(values)
		if err != nil {
			return
		}

		// The various travel destinations that the trip from A to B
		// has to go through. This is a list consisting of transportation
		// methods that is used. For example: {"38"} if the entire trip
		// consists of taking route 38.
		routeSpan := findStrict(mainContent, "span", "class", "tm-det-wrapper tm-alpha8")

		// Treat the children as an ordered list of edges from travel method
		// to travel method.
		travelDests := findAll(routeSpan, "span", "class", "tm-det")

		for _, td := range travelDests {
			var t Transport

			// Each td has a span 'tm-det-text-walk' and a span 'tm-det-transport' (containing
			// 'tm-det-linenr'). To determine if the transport method is walking or by bus
			// check if 'tm-det-text-walk' is non-empty, as it should contain "Walking"
			// if the given transport is walking, and empty if by bus (or other transport).
			walkSpan := findStrict(td, "span", "class", "tm-det-text tm-det-text-walk")

			walkText := walkSpan.Text()

			if walkText != "" {
				t.WalkText = walkText
				t.Type = TransportWalking
			} else {
				t.Type = TransportBus

				lineNumStr := findStrict(td, "span", "class", "tm-det-linenr").Text()
				t.LineNum, err = strconv.Atoi(lineNumStr)
				if err != nil {
					return
				}
			}

			// Each td also contains a span with the time 'tm-det-time'.
			detTimeSpan := findStrict(td, "span", "class", "ui-helper-hidden-accessible tm-det-time")
			t.Start, _ = dateTimeMerge(dateStr, detTimeSpan.Text())

			d.Route = append(d.Route, t)
		}
		deps = append(deps, d)
	}

	return deps, nil
}

func GetDeparturesNow(dir int, from, to string) ([]Departure, error) {
	now := time.Now()

	dtime := fmt.Sprintf("%02d:%02d", now.Hour(), now.Minute())
	ddate := fmt.Sprintf("%02d.%02d.%02d", now.Day(), now.Month(), now.Year())

	return GetDepartures(dir, from, to, dtime, ddate)
}
