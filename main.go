// AtB route planner cli in go.
// Date: 12.04.2019
// Author: Jørgen Bele Reinfjell
// Description: AtB route planner cli

package main

import (
	"encoding/json"
	"fmt"
	"github.com/b4b4r07/go-finder"
	"github.com/b4b4r07/go-finder/source"
	"github.com/docopt/docopt-go"
	"github.com/jorgenbele/go-atb/atb"
	"os"
	"strconv"
	"strings"
	"time"
)

func bold(s string) string {
	return fmt.Sprintf("\033[1m%s\033[0m", s)
}

func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func min(x, y int) int {
	if x > y {
		return y
	}
	return x
}

func printRow(row []string, cw []int, spacing int, bold bool) {
	if bold {
		fmt.Printf("\033[1m")
	}

	for i, c := range row[:len(row)-1] {
		fmt.Printf("%-*s%*c", cw[i], c[:min(cw[i], len(c))], spacing, ' ')
	}
	li := len(row) - 1
	lc := row[li]
	fmt.Printf("%-*s\n", cw[li], lc[:min(cw[li], len(lc))])

	if bold {
		fmt.Printf("\033[0m")
	}
}

func printPlanMinimal(deps []atb.Departure) {
	fmt.Printf("%-5s| %-3s |%3s|%4s|%1s|%-10s\n", "Start", "End", "Durat", "Chan", "F", "Route")
	for _, d := range deps {
		route := make([]string, len(d.Route))
		for i, r := range d.Route {
			if r.Type == atb.TransportBus {
				route[i] = strconv.Itoa(r.LineNum)
			} else {
				route[i] = r.WalkText
			}
		}
		routeStr := strings.Join(route, " ⟶  ")
		fmt.Printf("%02d:%02d|%02d:%02d|%3.0f m|%4d|%1s|%s\n", d.Start.Hour(), d.Start.Minute(), d.End.Hour(), d.End.Minute(),
			d.Duration.Minutes(), d.Changes, d.Fare, routeStr)
	}
}

const (
	// NoRoute is used when no route is provided as args.
	NoRoute = 0
)

func printRealtimeList(rdeps []atb.RealtimeDeparture, allowBold bool, route int) {
	header := []string{"ROUTE", "TIME", "TOWARDS", "REALTIME"}
	spacing := 2

	rows := make([][]string, 0, len(rdeps))
	cw := make([]int, len(header))
	bold := make([]bool, len(rdeps))

	for _, d := range rdeps {
		var rtStr string

		if route != NoRoute && d.Transport.Type == atb.TransportBus && d.Transport.LineNum != route {
			continue
		}

		if d.IsRealtime {
			rtStr = "TRUE"
			bold[len(rows)] = true
		} else {
			rtStr = "FALSE"
		}

		row := []string{
			fmt.Sprintf("%5d", d.Transport.LineNum),
			fmt.Sprintf("%02d:%02d", d.Transport.Start.Hour(), d.Transport.Start.Minute()),
			d.Towards,
			rtStr,
		}

		for j, c := range row {
			cw[j] = max(cw[j], len(c))
		}

		rows = append(rows, row)
	}

	printRow(header, cw, spacing, false)
	for i, r := range rows {
		printRow(r, cw, spacing, bold[i] && allowBold)
	}
}

type symbols struct {
	routeSplit string

	bus     string
	train   string
	tram    string
	walking string

	start    string
	end      string
	duration string
	changes  string
	fare     string

	time     string
	realtime string

	towards string
}

func printPlanTabular(deps []atb.Departure, symb symbols) {
	header := []string{"START", "END", "DURATION", "CHANGES", "FARE", "ROUTE"}
	spacing := 2

	rows := make([][]string, len(deps))
	cw := make([]int, len(header))

	for i, d := range deps {
		// Join routes into a presentable string.
		route := make([]string, len(d.Route))
		for i, r := range d.Route {
			if r.Type == atb.TransportBus {
				//route[i] = fmt.strconv.Itoa(r.LineNum)
				route[i] = fmt.Sprintf("%s%d", symb.bus, r.LineNum)
			} else {
				//route[i] = r.WalkText
				route[i] = fmt.Sprintf("%s%s", symb.walking, r.WalkText)
			}
		}
		routeStr := strings.Join(route, symb.routeSplit)

		rows[i] = []string{
			fmt.Sprintf("%s%02d:%02d", symb.start, d.Start.Hour(), d.Start.Minute()),
			fmt.Sprintf("%s%02d:%02d", symb.end, d.End.Hour(), d.End.Minute()),
			fmt.Sprintf("%s%vm", symb.duration, d.Duration.Minutes()),
			fmt.Sprintf("%s%d", symb.changes, d.Changes),
			fmt.Sprintf("%s%s", symb.fare, d.Fare),
			//string(d.Fare),
			routeStr,
		}

		for j, c := range rows[i] {
			cw[j] = max(cw[j], len(c))
		}
	}

	printRow(header, cw, spacing, false)
	for _, r := range rows {
		printRow(r, cw, spacing, false)
	}
}

func getSuggestions(from, to string) (selFrom, selTo string) {
	fromChan := make(chan []string)
	toChan := make(chan []string)

	f := func(query string, resChan chan []string) {
		v, err := atb.GetSuggestions(query)
		if err != nil {
			panic(fmt.Sprintf("Unable to get suggestion: %v", err))
		}
		resChan <- v
	}

	go f(from, fromChan)
	go f(to, toChan)

	sFrom := <-fromChan
	sTo := <-toChan

	// Lazy init the finder only when needed in case no finder
	// is installed and the init fails.
	var lazyfinder finder.Finder
	lazyfinder = nil

	userSelect := func(orig string, suggestions []string) string {
		switch len(suggestions) {
		case 0:
			return orig
		case 1:
			return suggestions[0]
		default:
			if lazyfinder == nil {
				var err error
				lazyfinder, err = finder.New()
				if err != nil {
					panic(err)
				}
			}

			lazyfinder.Read(source.Slice(suggestions))
			selected, err := lazyfinder.Run()
			if err != nil {
				panic(err)
			}
			// Take the first one, assume the user selected only one.
			return selected[0]
		}
	}

	selFrom = userSelect(from, sFrom)
	selTo = userSelect(to, sTo)
	return
}

func dumpJSON(v interface{}) {
	json, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("Unable to convert to json: %v", err))
	}
	os.Stdout.Write(json)
}

func main() {
	usage := `AtB Travel Planner

Usage: atb [--terse | --json] (([--minimal] <from> <to> [(--departure | --arrival) <time> [<date>]]) [--no-suggestions] | --suggestions <query> | --serve-api <apikey>)

Options:
       --terse                        Disables bold lines and use of symbols.

       --no-suggestions               Disables the use of the suggestions feature which does a lookup
                                      of stations with name <from> (and <to> if not --realtime). This
                                      is useful when you have the complete unique name of a station.

       --suggestions <query>          Does a station lookup using the string <query> and exits.

       --json                         Dump output in JSON format instead of in tabulated form.

       --minimal                      Use an alternative tabulated output format. CONFLICTS with --json.

       --departure <time> [<date>]    Depart at time <time> on date <date> (today if not specified).
       --arrival   <time> [<date>]    Arrive by time <time> on date <date> (today if not specified).

Formatting:
       <time> has to be on the format: HOUR:MINUTE
       <date> has to be on the format: DAYOFMONTH.MONTH.YEAR
`
	var opts docopt.Opts
	var err error

	var config struct {
		ToArg            string `docopt:"<to>"`
		FromArg          string `docopt:"<from>"`
		NoSuggestions    bool   `docopt:"--no-suggestions"`
		Query            string `docopt:"<query>"`
		OnlySuggestions  bool   `docopt:"--suggestions"`
		Terse            bool   `docopt:"--terse"`
		Route            int    `docopt:"<route>"`
		Minimal          bool   `docopt:"--minimal"`
		DumpJSON         bool   `docopt:"--json"`
		UseArrivalTime   bool   `docopt:"--arrival"`
		UseDepartureTime bool   `docopt:"--departure"`
		Date             string `docopt:"<date>"`
		Time             string `docopt:"<time>"`
		ServeAPI         bool   `docopt:"--serve-api"`
		ServeAPIKey      string `docopt:"<apikey>"`
	}

	if opts, err = docopt.ParseDoc(usage); err != nil {
		panic(fmt.Sprintf("Unable to parse arguments: %v\n", err))
	}

	opts.Bind(&config)

	if config.ServeAPI {
		r := atb.CreateAPIEngine(config.ServeAPIKey)
		r.Run()
		return
	}

	var symbs symbols
	if config.Terse {
		symbs = symbols{routeSplit: ","}
	} else {
		symbs = symbols{
			routeSplit: " ⟶  ",
			bus:        "🚍",
			train:      " ",
			tram:       " ",
			walking:    " ",
			//towards:      " ",
			//start:        " ",
			//end:          " ",
			//duration:     " ",
			//fare:         "",
			//changes:      "",
		}
	}

	if config.OnlySuggestions {
		v, err := atb.GetSuggestions(config.Query)
		if err != nil {
			panic(fmt.Sprintf("Unable to get suggestion: %v", err))
		}

		if config.DumpJSON {
			dumpJSON(v)
			return
		}

		for _, e := range v {
			fmt.Printf("%v\n", e)
		}
		return
	}

	var to, from string
	if config.NoSuggestions {
		from, to = config.FromArg, config.ToArg
	} else {
		// Get suggestions in parallel.
		from, to = getSuggestions(config.FromArg, config.ToArg)
	}

	var deps []atb.Departure
	if config.UseArrivalTime || config.UseDepartureTime {
		date := config.Date
		if date == "" {
			now := time.Now()
			date = fmt.Sprintf("%02d.%02d.%02d", now.Day(), now.Month(), now.Year())
		}
		deps, err = atb.GetDeparturesReq(atb.DepartureReq{From: from, To: to, Time: config.Time, Date: date, IsArrivalTime: config.UseArrivalTime, IsRealtime: false})
	} else {
		deps, err = atb.GetDeparturesNow(from, to)
	}

	if err != nil {
		// Panic to get more debugging output.
		panic(fmt.Sprintf("Unable to get departures: %v\n", err))
		//fmt.Fprintf(os.Stderr, "Error: Unable to get departures: %v\n", err)
		//os.Exit(1)
	}

	if config.DumpJSON {
		dumpJSON(deps)
		return
	}

	if !config.Terse {
		fmt.Printf(bold(":: From %s to %s\n"), from, to)
	}

	if config.Minimal {
		printPlanMinimal(deps)
		return
	}
	printPlanTabular(deps, symbs)
}
