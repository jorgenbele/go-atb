// AtB route planner cli in go.
// Date: 12.04.2019
// Author: Jørgen Bele Reinfjell
// Description: AtB route planner cli

package main

import (
	"fmt"
	"github.com/b4b4r07/go-finder"
	"github.com/b4b4r07/go-finder/source"
	"github.com/docopt/docopt-go"
	"github.com/jorgenbele/go-atb/atb"
	"strconv"
	"strings"
)

func bold(s string) string {
	return fmt.Sprintf("\033[1m%s\033[0m", s)
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

func printPlanTabular(deps []atb.Departure) {
	header := []string{"START", "END", "DURATION", "CHANGES", "FARE", "ROUTE"}
	spacing := 2

	rows := make([][]string, len(deps))
	cw := make([]int, len(header))

	for i, d := range deps {
		// Join routes into a presentable string.
		route := make([]string, len(d.Route))
		for i, r := range d.Route {
			if r.Type == atb.TransportBus {
				route[i] = strconv.Itoa(r.LineNum)
			} else {
				route[i] = r.WalkText
			}
		}
		routeStr := strings.Join(route, " ⟶  ")

		rows[i] = []string{
			fmt.Sprintf("%02d:%02d", d.Start.Hour(), d.Start.Minute()),
			fmt.Sprintf("%02d:%02d", d.End.Hour(), d.End.Minute()),
			fmt.Sprintf("%v m", d.Duration.Minutes()),
			fmt.Sprintf("%d", d.Changes),
			string(d.Fare),
			routeStr,
		}

		max := func(x, y int) int {
			if x > y {
				return x
			}
			return y
		}

		for j, c := range rows[i] {
			cw[j] = max(cw[j], len(c))
		}
	}

	// start :padding: end :padding: duration :padding: changes :padding: fare :padding: route
	format := "%-*s%*c %-*s%*c %-*s%*c %-*s%*c %-*s%*c %-*s\n"

	min := func(x, y int) int {
		if x > y {
			return y
		}
		return x
	}

	h := header
	fmt.Printf(format, cw[0], h[0][:min(cw[0], len(h[0]))], spacing-1, ' ',
		cw[1], h[1][:min(cw[1], len(h[1]))], spacing-1, ' ',
		cw[2], h[2][:min(cw[2], len(h[2]))], spacing-1, ' ',
		cw[3], h[3][:min(cw[3], len(h[3]))], spacing-1, ' ',
		cw[4], h[4][:min(cw[4], len(h[4]))], spacing-1, ' ',
		cw[5], h[5][:min(cw[5], len(h[5]))])

	for _, r := range rows {
		fmt.Printf(format, cw[0], r[0], spacing-1, ' ',
			cw[1], r[1], spacing-1, ' ',
			cw[2], r[2], spacing-1, ' ',
			cw[3], r[3], spacing-1, ' ',
			cw[4], r[4], spacing-1, ' ',
			cw[5], r[5])
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

	// Lazy init the finder only when needed.
	var finder_ finder.Finder
	finder_ = nil

	userSelect := func(orig string, suggestions []string) string {
		switch len(suggestions) {
		case 0:
			return orig
		case 1:
			return suggestions[0]
		default:
			if finder_ == nil {
				var err error
				finder_, err = finder.New()
				if err != nil {
					panic(err)
				}
			}

			finder_.Read(source.Slice(suggestions))
			selected, err := finder_.Run()
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

func main() {
	usage := `AtB Travel Planner

Usage: atb (<from> <to> [--no-suggestions]| --suggestions <query>)
`
	var opts docopt.Opts
	var err error

	var config struct {
		ToArg           string `docopt:"<to>"`
		FromArg         string `docopt:"<from>"`
		NoSuggestions   bool   `docopt:"--no-suggestions"`
		Query           string `docopt:"<query>"`
		OnlySuggestions bool   `docopt:"--suggestions"`
	}

	if opts, err = docopt.ParseDoc(usage); err != nil {
		panic(fmt.Sprintf("Unable to parse arguments: %v\n", err))
	}

	opts.Bind(&config)

	if config.OnlySuggestions {
		v, err := atb.GetSuggestions(config.Query)
		if err != nil {
			panic(fmt.Sprintf("Unable to get suggestion: %v", err))
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

	deps, err := atb.GetDeparturesNow(1, from, to)
	if err != nil {
		panic(fmt.Sprintf("Unable to get departures: %v\n", err))
		//fmt.Fprintf(os.Stderr, "Error: Unable to get departures: %v\n", err)
		//os.Exit(1)
	}

	fmt.Printf(bold(":: From %s to %s\n"), from, to)
	//printPlanMinimal(deps)
	printPlanTabular(deps)
}
