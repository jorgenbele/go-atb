package main

import (
	"fmt"
	"github.com/jorgenbele/go-atb/atb"
	"os"
	"testing"
)

var realtimeDeps []atb.RealtimeDeparture
var deps []atb.Departure

func TestMain(m *testing.M) {
	var err error
	realtimeDeps, err = atb.GetRealtimeDepartures(1, "Solsiden (Trondheim)")
	if err != nil {
		fmt.Printf("Unable to get realtime departures: %v\n", err)
		os.Exit(1)
	}
	deps, err = atb.GetDeparturesNow(1, "Solsiden (Trondheim)", "Munkegata (Trondheim)")
	if err != nil {
		fmt.Printf("Unable to get departures: %v\n", err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

type tfunc func()

func toDevNull(f tfunc) {
	// XXX - Hack: redirect to /dev/null, should probably change printRealtimeList instead.
	// Keep backup of the real stdout.
	old := os.Stdout
	devnull, _ := os.OpenFile("/dev/null", os.O_RDWR, 0)
	os.Stdout = devnull

	f()

	// Restoring the real stdout.
	devnull.Close()
	os.Stdout = old
}

func BenchmarkPrintRealtime(b *testing.B) {
	toDevNull(func() {
		for n := 0; n < b.N; n++ {
			printRealtimeList(realtimeDeps, true, NoRoute)
		}
	})
}

func BenchmarkPrintPlanMinimal(b *testing.B) {
	toDevNull(func() {
		for n := 0; n < b.N; n++ {
			printPlanMinimal(deps)
		}
	})
}

func BenchmarkPrintPlanTabular(b *testing.B) {
	toDevNull(func() {
		for n := 0; n < b.N; n++ {
			printPlanTabular(deps, symbols{})
		}
	})
}
