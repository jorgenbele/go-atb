// AtB route planner cli in go.
// Date: 12.04.2019
// Author: Jørgen Bele Reinfjell
// Description: AtB route plannner cli to get next departure
//

package main

import (
	"fmt"
	"github.com/anaskhan96/soup"
	"github.com/levigross/grequests"
    "github.com/b4b4r07/go-finder/source"
    "github.com/b4b4r07/go-finder"
	"strconv"
	"strings"
	"time"
    "os"
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
	html, err := getDeparturesResp(dir, from, to, dtime, ddate)
	if err != nil {
		return
	}

	doc := soup.HTMLParse(html)
	mainContent := doc.FindStrict("div", "class", "maincontent")

	// Used when parsing start and end times.
	date := doc.FindStrict("h2", "class", "tm-alpha tm-reiseforslag-header").Text()
	resultWrappers := mainContent.FindAllStrict("div", "class", "tm-result-wrapper")

	for _, rw := range resultWrappers {
		var d Departure
		// The tm-block-b span contains:
		//     start and end time, duration, changes and fare
		blockB := rw.FindStrict("span", "class", "tm-block-b")

		// tm-block-b contains two tm-result-time-wrapper elements
		// where the first one is the start time, and the second
		// one is the end time.
		unpackTime := func(res []soup.Root) (start, end time.Time) {
			// Use the date of departure together with the start/end time
			// to convert to a time.Time object.
			start, err := dateTimeMerge(date, res[0].Text())
			if err != nil {
				fmt.Println(err)
			}
			end, err = dateTimeMerge(date, res[1].Text())
			if err != nil {
				fmt.Println(err)
			}
			return
		}
		d.Start, d.End = unpackTime(blockB.FindAll("span", "class", "tm-result-fratil"))

		// tm-result-details-extra contains three tm-result-info elements:
		// 1. duration, 2. changes, 3. fare
		// The 'tm-result-info-val' element contains the string describing each
		// element previously described.
		extraDetails := blockB.FindStrict("span", "class", "tm-inline-block tm-result-details-extra")
		spanClasses := []string{"tm-result-value-time", "tm-result-value-change", "tm-result-value-price"}
		values := [3]soup.Root{}
		for i, className := range spanClasses {
			values[i] = extraDetails.Find("span", "class", className).Find("span", "class", "tm-result-info-val")
		}

		parseDuration := func(duration string) (d time.Duration, err error) {
			durationSplice := strings.Split(duration, ":")
			hours, err := strconv.Atoi(durationSplice[0])
			if err != nil {
				return
			}
			minutes, err := strconv.Atoi(durationSplice[1])
			if err != nil {
				return
			}
			d = time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute
			return
		}

		unpackExtraDetails := func(res [3]soup.Root) (duration time.Duration, changes int, fare string) {
			duration, _ = parseDuration(res[0].Text()) // ignore error
			changes, _ = strconv.Atoi(res[1].Text())   // ignore error
			fare = res[2].Text()
			return
		}

		d.Duration, d.Changes, d.Fare = unpackExtraDetails(values)

		// The various travel destinations that the trip from A to B
		// has to go through. This is a list consisting of transportation
		// methods that is used. For example: {"38"} if the entire trip
		// consists of taking route 38.
		routeSpan := mainContent.FindStrict("span", "class", "tm-det-wrapper tm-alpha8")

		// Treat the children as an ordered list of edges from travel method
		// to travel method.
		travelDests := routeSpan.FindAll("span", "class", "tm-det")

		for _, td := range travelDests {
			var t Transport

			// Each td has a span 'tm-det-text-walk' and a span 'tm-det-transport' (containing
			// 'tm-det-linenr'). To determine if the transport method is walking or by bus
			// check if 'tm-det-text-walk' is non-empty, as it should contain "Walking"
			// if the given transport is walking, and empty if by bus (or other transport).
			walkText := td.FindStrict("span", "class", "tm-det-text tm-det-text-walk").Text()
			if walkText != "" {
				t.WalkText = walkText
				t.Type = TransportWalking
			} else {
				t.Type = TransportBus

				lineNumStr := td.FindStrict("span", "class", "tm-det-linenr").Text()
				t.LineNum, err = strconv.Atoi(lineNumStr)
				if err != nil {
					panic("Illegal line number")
				}
			}

			// Each td also contains a span with the time 'tm-det-time'.
			detTimeSpan := td.FindStrict("span", "class", "ui-helper-hidden-accessible tm-det-time")
			t.Start, _ = dateTimeMerge(date, detTimeSpan.Text())

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

func printPlanMinimal(deps []Departure) {
    fmt.Printf("%-5s| %-3s |%3s|%4s|%1s|%-10s\n", "Start", "End", "Durat", "Chan", "F", "Route")
    for _, d := range deps {
        route := make([]string, len(d.Route))
        for _, r := range d.Route {
            if r.Type == TransportBus {
                route = append(route, strconv.Itoa(r.LineNum))
            } else {
                route = append(route, r.WalkText)
            }
        }
        routeStr := route[0] + strings.Join(route[1:], " ⟶  ")
        fmt.Printf("%02d:%02d|%02d:%02d|%3.0f m|%4d|%1s|%s\n", d.Start.Hour(), d.Start.Minute(), d.End.Hour(), d.End.Minute(), d.Duration.Minutes(), d.Changes, d.Fare, routeStr)
    }
}

func main() {
    fromArg := os.Args[1]
    toArg := os.Args[2]

    // Get suggestions in parallell.
    fromChan := make(chan []string)
    toChan := make(chan []string)

    go func() {
        v, _ := GetSuggestions(fromArg)
        //fmt.Println("Got from:", v)
        fromChan <- v
    }()

    go func() {
        v, _ := GetSuggestions(toArg)
        //fmt.Println("Got to:", v)
        toChan <- v
    }()

    sFrom := <-fromChan
    sTo := <-toChan

    // TODO: Add cli flag.
    finder, err := finder.New()
    if err != nil {
        panic(err)
    }

	//fmt.Println(sFrom, sTo)
    var to, from string
    if len(sFrom) < 2 {
        from = sFrom[0]
    } else {
        finder.Read(source.Slice(sFrom))
        toSlice, err := finder.Run()
        if err != nil {
            panic(err)
        }
        from = toSlice[0]
    }


    if len(sTo) < 2 {
        to = sTo[0]
    } else {
        finder.Read(source.Slice(sTo))
        fromSlice, err := finder.Run()
        if err != nil {
            panic(err)
        }
        from = fromSlice[0]
    }

	deps, _ := GetDeparturesNow(1, from, to)
	//fmt.Println(deps)

	fmt.Printf("From %s to %s\n", from, to)
    printPlanMinimal(deps)


	//now := time.Now()
	//dtime := fmt.Sprintf("%02d:%02d", now.Hour(), now.Minute())
	//ddate := fmt.Sprintf("%02d.%02d.%02d", now.Day(), now.Month(), now.Year())
	//getDeparturesResp(1, "Konglevegen (Trondheim)", "Munkegata (Trondheim)", dtime, ddate)
}
