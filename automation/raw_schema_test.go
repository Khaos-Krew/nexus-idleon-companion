package main

import (
	"strings"
	"testing"
)

func TestDecodeToolboxProfileContract(t *testing.T) {
	raw := `{
	  "data": {
	    "Lv0_0": "[421]",
	    "CharacterClass_0": "16",
	    "CurrentMap_0": "130",
	    "CauldronBubbles": "[1,2,3,0]",
	    "StampLevels": "[10,20,0]",
	    "Construction": "{\"buildRate\":12345}",
	    "Sailing": "{\"artifacts\":[1,1,0]}"
	  },
	  "charNames": ["TestBubo"],
	  "serverVars": {"x":1},
	  "lastUpdated": 1770000000000
	}`
	snap, err := decodeFlexibleSnapshot([]byte(raw), "toolbox-test")
	if err != nil { t.Fatal(err) }
	if snap.Schema != "idleon-toolbox-profile" { t.Fatalf("schema=%q", snap.Schema) }
	if len(snap.Characters) != 1 { t.Fatalf("characters=%d", len(snap.Characters)) }
	if snap.Characters[0].Name != "TestBubo" || snap.Characters[0].Level != 421 { t.Fatalf("character=%+v", snap.Characters[0]) }
	for _, id := range []string{"alchemy","stamps","construction","sailing"} {
		if _, ok := snap.Systems[id]; !ok { t.Fatalf("expected %s detection", id) }
	}
}

func TestDecodeEfficiencyRawContract(t *testing.T) {
	raw := `{
	  "Lv0_0": [300],
	  "CharacterClass_0": 17,
	  "CurrentMap_0": 100,
	  "playerNames": ["TestES"],
	  "companions": [1,2],
	  "servervars": {"foo":"bar"},
	  "Farming": {"crops":[1,2,0]},
	  "Sneaking": {"jade":100},
	  "Summoning": {"wins":[1,1,0]},
	  "Research": {"observations":[1,0,0]},
	  "Spelunking": {"amber":15},
	  "SushiStation": {"sushi":[2,1,0]}
	}`
	snap, err := decodeFlexibleSnapshot([]byte(raw), "efficiency-test")
	if err != nil { t.Fatal(err) }
	if snap.Schema != "idleon-efficiency-raw" { t.Fatalf("schema=%q", snap.Schema) }
	if len(snap.Characters) != 1 || snap.Characters[0].Name != "TestES" { t.Fatalf("characters=%+v", snap.Characters) }
	if snap.World < 7 { t.Fatalf("world=%d expected inferred W7", snap.World) }
	for _, id := range []string{"farming","sneaking","summoning","research","spelunking","sushistation"} {
		state, ok := snap.Systems[id]
		if !ok { t.Fatalf("expected %s", id) }
		if strings.TrimSpace(state.Evidence) == "" { t.Fatalf("%s missing evidence", id) }
	}
}

func TestUnknownJSONFailsClosedToLowConfidenceDetection(t *testing.T) {
	raw := `{"Alchemy":{"bubbles":[1,0,0]},"characters":[{"name":"A","class":"Bubonic Conjuror","level":200}]}`
	snap, err := decodeFlexibleSnapshot([]byte(raw), "generic-test")
	if err != nil { t.Fatal(err) }
	state, ok := snap.Systems["alchemy"]
	if !ok { t.Fatal("alchemy not detected") }
	if state.Confidence > .70 { t.Fatalf("confidence too high: %v", state.Confidence) }
}
