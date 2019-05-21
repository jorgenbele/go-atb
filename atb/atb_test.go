package atb_test

import (
	"github.com/jorgenbele/go-atb/atb"
	"testing"
	//"fmt"
	//"time"
)

func suggestionsEqual(t *testing.T, query string, expected map[string]bool) {
	res, err := atb.GetSuggestions(query)
	if err != nil {
		t.Errorf("Unable to get suggestions for query: %s, %v", query, err)
	}

	m := make(map[string]bool)

	for _, r := range res {
		m[r] = true
		if _, ok := expected[r]; !ok {
			t.Errorf("query: %s, key %s not expected", query, r)
		}
	}

	for k := range expected {
		if _, ok := m[k]; !ok {
			t.Errorf("query: %s, expected key %s", query, k)
		}
	}
}

func TestGetSuggestions(t *testing.T) {
	suggestionsEqual(t, "munke", map[string]bool{
		"Munkegata (Trondheim)":    true,
		"Munkeby (Levanger)":       true,
		"Munkebykorsen (Levanger)": true,
		"Munken (Leksvik)":         true,
		"Munkebyvegen (Levanger)":  true,
		"Munken (Indre Fosen)":     true,
		"Munken (Ørland)":          true,
	})

	suggestionsEqual(t, "solsiden", map[string]bool{
		"Solsiden (Trondheim)": true,
	})

	suggestionsEqual(t, "MunKegata", map[string]bool{
		"Munkegata (Trondheim)": true,
	})

	suggestionsEqual(t, "samfundet", map[string]bool{
		"Samfundet (Trondheim)":            true,
		"Studentersamfundet (Trondheim)":   true,
		"Studentersamfundet 2 (Trondheim)": true,
	})
}

//func departuresEqual(t *testing.T, dir int, from, to string, expected []atb.Departure) {
//	// In the past.
//	dtime := "10:00"
//	ddate := "16.04.19"
//
//	resp, err := atb.GetDepartures(dir, from, to, dtime, ddate)
//	if err != nil {
//		t.Errorf("Unable to get depatrues from %s to %s at %s, %s in direction %d, %v", from, to, dtime, ddate, dir, err)
//	}
//
//	m := make(map[string]bool)
//
//	for i, r := range resp {
//		s := fmt.Sprintf("%v", r)
//		m[s] = true
//		e := fmt.Sprintf("%v", expected[i])
//		if e != s {
//			t.Errorf("not equal: %s and %s", e, s)
//		}
//	}
//}
//
//func TestGetDepartures(t *testing.T) {
//	parse := func(timestamp string) (tm time.Time) {
//		layout := "2006-01-02 15:04:05" //"Mon Jan 2 15:04:05 -0700 MST 2006"
//
//		tm, err := time.Parse(layout, timestamp)
//		if err != nil {
//			t.Errorf("Unable to parse timestamp: %s\n", timestamp)
//		}
//		return
//	}
//
//	parseDur := func(duration string) (dt time.Duration) {
//		dt, err := time.ParseDuration(duration)
//		if err != nil {
//			t.Errorf("Unable to parse duration: %s\n", duration)
//		}
//		return
//	}
//	departuresEqual(t, 1, "Munkegata (Trondheim)", "Solsiden (Trondheim)", []atb.Departure{
//		atb.Departure{parse("2019-04-16 10:01:00"), parse("2019-04-16 10:05:00"), 0, "-", parseDur("4m0s"), []atb.Transport{atb.Transport{atb.TransportBus, "", 9, parse("2019-04-16 10:01:00")}}},
//		atb.Departure{parse("2019-04-16 10:10:00"), parse("2019-04-16 10:14:00"), 0, "-", parseDur("4m0s"), []atb.Transport{atb.Transport{atb.TransportBus, "", 9, parse("2019-04-16 10:01:00")}}},
//		atb.Departure{parse("2019-04-16 10:11:00"), parse("2019-04-16 10:15:00"), 0, "-", parseDur("4m0s"), []atb.Transport{atb.Transport{atb.TransportBus, "", 9, parse("2019-04-16 10:01:00")}}},
//		atb.Departure{parse("2019-04-16 10:12:00"), parse("2019-04-16 10:16:00"), 0, "-", parseDur("4m0s"), []atb.Transport{atb.Transport{atb.TransportBus, "", 9, parse("2019-04-16 10:01:00")}}},
//		atb.Departure{parse("2019-04-16 10:13:00"), parse("2019-04-16 10:17:00"), 0, "-", parseDur("4m0s"), []atb.Transport{atb.Transport{atb.TransportBus, "", 9, parse("2019-04-16 10:01:00")}}},
//		atb.Departure{parse("2019-04-16 10:15:00"), parse("2019-04-16 10:19:00"), 0, "-", parseDur("4m0s"), []atb.Transport{atb.Transport{atb.TransportBus, "", 9, parse("2019-04-16 10:01:00")}}},
//		atb.Departure{parse("2019-04-16 10:16:00"), parse("2019-04-16 10:20:00"), 0, "-", parseDur("4m0s"), []atb.Transport{atb.Transport{atb.TransportBus, "", 9, parse("2019-04-16 10:01:00")}}},
//		atb.Departure{parse("2019-04-16 10:21:00"), parse("2019-04-16 10:25:00"), 0, "-", parseDur("4m0s"), []atb.Transport{atb.Transport{atb.TransportBus, "", 9, parse("2019-04-16 10:01:00")}}},
//		atb.Departure{parse("2019-04-16 10:27:00"), parse("2019-04-16 10:31:00"), 0, "-", parseDur("4m0s"), []atb.Transport{atb.Transport{atb.TransportBus, "", 9, parse("2019-04-16 10:01:00")}}},
//		atb.Departure{parse("2019-04-16 10:30:00"), parse("2019-04-16 10:34:00"), 0, "-", parseDur("4m0s"), []atb.Transport{atb.Transport{atb.TransportBus, "", 9, parse("2019-04-16 10:01:00")}}},
//	})
//
//	departuresEqual(t, 1, "Hommelvik stasjon (Malvik)", "Studentersamfundet (Trondheim)", []atb.Departure{
//		atb.Departure{parse("2019-04-16 10:03:00"), parse("2019-04-16 10:42:00"), 1, "-", parseDur("39m0s"), []atb.Transport{atb.Transport{atb.TransportBus, "", 26, parse("2019-04-16 10:03:00")}, atb.Transport{atb.TransportWalking, "Walk", 0, parse("2019-04-16 10:32:00")}, atb.Transport{atb.TransportBus, "", 55, parse("2019-04-16 10:35:00")}}},
//		atb.Departure{parse("2019-04-16 10:06:00"), parse("2019-04-16 10:55:00"), 0, "-", parseDur("49m0s"), []atb.Transport{atb.Transport{atb.TransportBus, "", 26, parse("2019-04-16 10:03:00")}, atb.Transport{atb.TransportWalking, "Walk", 0, parse("2019-04-16 10:32:00")}, atb.Transport{atb.TransportBus, "", 55, parse("2019-04-16 10:35:00")}}},
//		atb.Departure{parse("2019-04-16 10:06:00"), parse("2019-04-16 10:48:00"), 1, "-", parseDur("42m0s"), []atb.Transport{atb.Transport{atb.TransportBus, "", 26, parse("2019-04-16 10:03:00")}, atb.Transport{atb.TransportWalking, "Walk", 0, parse("2019-04-16 10:32:00")}, atb.Transport{atb.TransportBus, "", 55, parse("2019-04-16 10:35:00")}}},
//		atb.Departure{parse("2019-04-16 10:36:00"), parse("2019-04-16 11:25:00"), 0, "-", parseDur("49m0s"), []atb.Transport{atb.Transport{atb.TransportBus, "", 26, parse("2019-04-16 10:03:00")}, atb.Transport{atb.TransportWalking, "Walk", 0, parse("2019-04-16 10:32:00")}, atb.Transport{atb.TransportBus, "", 55, parse("2019-04-16 10:35:00")}}},
//		atb.Departure{parse("2019-04-16 10:36:00"), parse("2019-04-16 11:20:00"), 1, "-", parseDur("44m0s"), []atb.Transport{atb.Transport{atb.TransportBus, "", 26, parse("2019-04-16 10:03:00")}, atb.Transport{atb.TransportWalking, "Walk", 0, parse("2019-04-16 10:32:00")}, atb.Transport{atb.TransportBus, "", 55, parse("2019-04-16 10:35:00")}}},
//		atb.Departure{parse("2019-04-16 11:03:00"), parse("2019-04-16 11:42:00"), 1, "-", parseDur("39m0s"), []atb.Transport{atb.Transport{atb.TransportBus, "", 26, parse("2019-04-16 10:03:00")}, atb.Transport{atb.TransportWalking, "Walk", 0, parse("2019-04-16 10:32:00")}, atb.Transport{atb.TransportBus, "", 55, parse("2019-04-16 10:35:00")}}},
//		atb.Departure{parse("2019-04-16 11:06:00"), parse("2019-04-16 11:55:00"), 0, "-", parseDur("49m0s"), []atb.Transport{atb.Transport{atb.TransportBus, "", 26, parse("2019-04-16 10:03:00")}, atb.Transport{atb.TransportWalking, "Walk", 0, parse("2019-04-16 10:32:00")}, atb.Transport{atb.TransportBus, "", 55, parse("2019-04-16 10:35:00")}}},
//		atb.Departure{parse("2019-04-16 11:06:00"), parse("2019-04-16 11:48:00"), 1, "-", parseDur("42m0s"), []atb.Transport{atb.Transport{atb.TransportBus, "", 26, parse("2019-04-16 10:03:00")}, atb.Transport{atb.TransportWalking, "Walk", 0, parse("2019-04-16 10:32:00")}, atb.Transport{atb.TransportBus, "", 55, parse("2019-04-16 10:35:00")}}},
//		atb.Departure{parse("2019-04-16 11:36:00"), parse("2019-04-16 12:25:00"), 0, "-", parseDur("49m0s"), []atb.Transport{atb.Transport{atb.TransportBus, "", 26, parse("2019-04-16 10:03:00")}, atb.Transport{atb.TransportWalking, "Walk", 0, parse("2019-04-16 10:32:00")}, atb.Transport{atb.TransportBus, "", 55, parse("2019-04-16 10:35:00")}}},
//		atb.Departure{parse("2019-04-16 11:36:00"), parse("2019-04-16 12:20:00"), 1, "-", parseDur("44m0s"), []atb.Transport{atb.Transport{atb.TransportBus, "", 26, parse("2019-04-16 10:03:00")}, atb.Transport{atb.TransportWalking, "Walk", 0, parse("2019-04-16 10:32:00")}, atb.Transport{atb.TransportBus, "", 55, parse("2019-04-16 10:35:00")}}},
//	})
//}
//
//// I dont know how to test realtime.
