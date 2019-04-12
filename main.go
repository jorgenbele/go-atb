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

func main() {
    usage := `AtB Travel Planner

Usage: atb <from> <to>`

    args, err := docopt.ParseDoc(usage)
    if err != nil {
        panic(fmt.Sprintf("Unable to parse arguments: %v\n", err))
    }

	fromArg, err := args.String("<from>")
    if err != nil {
        panic(err)
    }

	toArg, err := args.String("<to>")
    if err != nil {
        panic(err)
    }

	// Get suggestions in parallell.
	fromChan := make(chan []string)
	toChan := make(chan []string)

	go func() {
		v, err := atb.GetSuggestions(fromArg)
		if err != nil {
			panic(fmt.Sprintf("Unable to get suggestion: %v", err))
		}
		fromChan <- v
	}()

	go func() {
		v, err := atb.GetSuggestions(toArg)
		if err != nil {
			panic(fmt.Sprintf("Unable to get suggestion: %v", err))
		}
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
		toSlice, err := finder.Run()
		if err != nil {
			panic(err)
		}
		to = toSlice[0]
	}

	deps, _ := atb.GetDeparturesNow(1, from, to)

	fmt.Printf(bold(":: From %s to %s\n"), from, to)
	printPlanMinimal(deps)
}
