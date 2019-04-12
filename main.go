// AtB route planner cli in go.
// Date: 12.04.2019
// Author: Jørgen Bele Reinfjell
// Description: AtB route planner cli

// TODO: Support boats, etc.

package main

import (
	"fmt"
	"github.com/b4b4r07/go-finder"
	"github.com/b4b4r07/go-finder/source"
	"github.com/docopt/docopt-go"
	//"github.com/jorgenbele/go-atb/atb"
	"go-atb/atb"
	"os"
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

Usage: atb <from> <to> [options]

Options:
    --no-suggestions    Always send 'from' and 'to' as provided.
`
	var opts docopt.Opts
	var err error

	var config struct {
		ToArg         string `docopt:"<to>"`
		FromArg       string `docopt:"<from>"`
		NoSuggestions bool   `docopt:"--no-suggestions"`
	}

	if opts, err = docopt.ParseDoc(usage); err != nil {
		panic(fmt.Sprintf("Unable to parse arguments: %v\n", err))
	}

	opts.Bind(&config)
	fmt.Println(config)

	var to, from string
	// Get suggestions in parallell.
	if config.NoSuggestions {
		from, to = config.FromArg, config.ToArg
	} else {
		from, to = getSuggestions(config.FromArg, config.ToArg)
	}

	deps, err := atb.GetDeparturesNow(1, from, to)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Unable to get departures: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf(bold(":: From %s to %s\n"), from, to)
	printPlanMinimal(deps)
}
