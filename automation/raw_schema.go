package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// rawSystemSignatures intentionally contains both the friendly community-tool
// names and stable-ish save-key fragments seen in Toolbox/Efficiency raw data.
// Matching a signature proves presence, not correctness of a calculated bonus.
var rawSystemSignatures = map[string][]string{
	"worldpush": {"CurrentMap", "Portal", "UnlockedMaps", "MapUnlocked"},
	"stamps": {"StampLevel", "StampLv", "Stamps"},
	"vault": {"Vault", "UpgradeVault"},
	"forge": {"Forge", "ForgeSlots"},
	"anvil": {"AnvilPA", "Anvil", "Smithing"},
	"statues": {"StatueLevels", "Statues"},
	"cards": {"Cards", "CardSet", "CardLevels"},
	"constellations": {"StarSign", "Constellation"},
	"talents": {"Talent", "SkillPreset"},
	"gear": {"EquipOrder", "EquipmentOrder", "EquipmentMap"},
	"quests": {"Quest", "Quests"},
	"tasks": {"Task", "Merit"},
	"dungeons": {"Dungeon", "Flurbo"},
	"alchemy": {"Cauldron", "Bubble", "Vial", "Alchemy"},
	"prisma": {"Prisma"},
	"postoffice": {"POu_", "PostOffice", "PostOfficeInfo"},
	"obols": {"Obol"},
	"refinery": {"Refinery", "Salt"},
	"construction": {"Construction", "BuildRate", "Tower"},
	"printer": {"Printer", "3DPrinter", "Sample"},
	"worship": {"Worship", "Soul", "Totem"},
	"trapping": {"Trap", "Trapping"},
	"shrines": {"Shrine"},
	"deathnote": {"DeathNote", "Death Note", "MonsterKill"},
	"cooking": {"Cooking", "Meal", "Kitchen"},
	"breeding": {"Breeding", "Pet", "Territory"},
	"lab": {"Laboratory", "Lab", "Jewel"},
	"rift": {"Rift"},
	"tome": {"Tome"},
	"killroy": {"KillRoy", "Killroy"},
	"divinity": {"Divinity", "Deity", "God"},
	"sailing": {"Sailing", "Artifact", "Boat"},
	"gaming": {"Gaming", "Superbit"},
	"companions": {"companions", "Companion"},
	"hole": {"TheHole", "Hole", "Cavern", "Villager"},
	"slab": {"Slab", "Looty"},
	"sneaking": {"Sneaking", "Ninja", "Jade"},
	"farming": {"Farming", "Crop", "LandRank"},
	"summoning": {"Summoning", "Slime"},
	"beanstalk": {"Beanstalk"},
	"masterclasses": {"MasterClass", "Master Class"},
	"research": {"Research", "Observation"},
	"spelunking": {"Spelunk", "Tunnel", "Amber"},
	"coralreef": {"CoralReef", "Coral", "Fishies"},
	"sushistation": {"SushiStation", "Sushi", "Bucks"},
	"greenstacks": {"GreenStack", "Green Stack"},
	"bosses": {"Boss", "BossKey"},
	"minibosses": {"MiniBoss", "Miniboss"},
	"colosseum": {"Colosseum", "Colo"},
	"weeklyboss": {"WeeklyBoss", "Weekly Boss"},
	"vman": {"Voidwalker", "Speedrun"},
	"owl": {"Owl"},
	"roo": {"Kangaroo", "Roo"},
}

var slotKeyRE = regexp.MustCompile(`^(?:Lv0|CharacterClass|CurrentMap|AFKtarget|PTimeAway|SL|SM|PVStatList)_(\d+)$`)

func parseJSONish(v any) any {
	s, ok := v.(string)
	if !ok { return v }
	t := strings.TrimSpace(s)
	if t == "" { return v }
	if !(strings.HasPrefix(t, "{") || strings.HasPrefix(t, "[") || t == "true" || t == "false" || t == "null" || strings.HasPrefix(t, "\"") || (t[0] >= '0' && t[0] <= '9') || t[0] == '-') { return v }
	var out any
	dec := json.NewDecoder(strings.NewReader(t))
	dec.UseNumber()
	if err := dec.Decode(&out); err == nil { return normalizeJSONNumbers(out) }
	return v
}

func normalizeJSONNumbers(v any) any {
	switch t := v.(type) {
	case json.Number:
		if f, err := t.Float64(); err == nil { return f }
		return t.String()
	case map[string]any:
		for k, x := range t { t[k] = normalizeJSONNumbers(x) }
		return t
	case []any:
		for i, x := range t { t[i] = normalizeJSONNumbers(x) }
		return t
	default:
		return v
	}
}

func deepParseJSONish(v any) any {
	v = parseJSONish(v)
	switch t := v.(type) {
	case map[string]any:
		for k, x := range t { t[k] = deepParseJSONish(x) }
	case []any:
		for i, x := range t { t[i] = deepParseJSONish(x) }
	}
	return v
}

func detectCommunitySchema(root map[string]any) string {
	if _, ok := root["systems"]; ok {
		if _, chars := root["characters"]; chars { return "normalized-agent-snapshot" }
	}
	_, hasData := root["data"]
	_, hasCharNames := root["charNames"]
	_, hasServerVars := root["serverVars"]
	if hasData && (hasCharNames || hasServerVars) { return "idleon-toolbox-profile" }
	_, hasPlayerNames := root["playerNames"]
	_, hasServervars := root["servervars"]
	if hasPlayerNames || hasServervars { return "idleon-efficiency-raw" }
	if hasData { return "toolbox-legacy-profile" }
	return "generic-json"
}

func unwrapCommunityPayload(root map[string]any) (map[string]any, []string, string, []string) {
	schema := detectCommunitySchema(root)
	warnings := []string{}
	names := extractStringArray(root, "charNames", "playerNames")

	if schema == "idleon-toolbox-profile" || schema == "toolbox-legacy-profile" {
		if raw, ok := root["data"].(map[string]any); ok {
			parsed := deepParseJSONish(raw).(map[string]any)
			// Keep the profile metadata alongside the game data using the same names
			// Efficiency exposes on its Raw Data page.
			if len(names) > 0 { parsed["playerNames"] = stringsToAny(names) }
			if v, ok := root["companion"]; ok { parsed["companions"] = deepParseJSONish(v) }
			if v, ok := root["serverVars"]; ok { parsed["servervars"] = deepParseJSONish(v) }
			if v, ok := root["guildData"]; ok { parsed["guildData"] = deepParseJSONish(v) }
			if v, ok := root["tournament"]; ok { parsed["tournament"] = deepParseJSONish(v) }
			return parsed, names, schema, warnings
		}
		warnings = append(warnings, "Toolbox profile was detected but its data field was not an object")
	}
	return deepParseJSONish(root).(map[string]any), names, schema, warnings
}

func stringsToAny(in []string) []any { out := make([]any, len(in)); for i, s := range in { out[i] = s }; return out }

func extractStringArray(root map[string]any, keys ...string) []string {
	for _, wanted := range keys {
		for k, v := range root {
			if !strings.EqualFold(k, wanted) { continue }
			v = deepParseJSONish(v)
			a, ok := v.([]any); if !ok { continue }
			var out []string
			for _, x := range a { if s, ok := x.(string); ok && strings.TrimSpace(s) != "" { out = append(out, strings.TrimSpace(s)) } }
			if len(out) > 0 { return out }
		}
	}
	return nil
}

func valueAtKey(root map[string]any, key string) (any, bool) {
	for k, v := range root { if strings.EqualFold(k, key) { return deepParseJSONish(v), true } }
	return nil, false
}

func firstNumeric(v any) (float64, bool) {
	v = deepParseJSONish(v)
	switch t := v.(type) {
	case float64: return t, true
	case int: return float64(t), true
	case json.Number: f, e := t.Float64(); return f, e == nil
	case string: f, e := strconv.ParseFloat(strings.TrimSpace(t), 64); return f, e == nil
	case []any:
		for _, x := range t { if n, ok := firstNumeric(x); ok { return n, true } }
	case map[string]any:
		for _, key := range []string{"level", "progress", "value", "amount", "count"} { if x, ok := valueAtKey(t, key); ok { if n, ok := firstNumeric(x); ok { return n, true } } }
	}
	return 0, false
}

func classLabelFromRaw(v any) (string, int) {
	if s, ok := v.(string); ok {
		s = strings.TrimSpace(s)
		if s != "" { if n, err := strconv.Atoi(s); err == nil { return fmt.Sprintf("Class #%d", n), n }; return strings.ReplaceAll(s, "_", " "), 0 }
	}
	if n, ok := firstNumeric(v); ok { return fmt.Sprintf("Class #%d", int(n)), int(n) }
	return "", 0
}

func extractRawCharacters(raw map[string]any, names []string) []CharacterSnapshot {
	slots := map[int]bool{}
	for k := range raw { if m := slotKeyRE.FindStringSubmatch(k); len(m) == 2 { if n, err := strconv.Atoi(m[1]); err == nil { slots[n] = true } } }
	if len(slots) == 0 && len(names) > 0 { for i := range names { slots[i] = true } }
	var ids []int; for n := range slots { ids = append(ids, n) }; sort.Ints(ids)
	out := make([]CharacterSnapshot, 0, len(ids))
	for _, id := range ids {
		c := CharacterSnapshot{Name: fmt.Sprintf("Character %d", id+1)}
		if id < len(names) && names[id] != "" { c.Name = names[id] }
		if v, ok := valueAtKey(raw, fmt.Sprintf("Lv0_%d", id)); ok { if n, ok := firstNumeric(v); ok { c.Level = int(n) } }
		if v, ok := valueAtKey(raw, fmt.Sprintf("CharacterClass_%d", id)); ok { c.Class, c.ClassIndex = classLabelFromRaw(v) }
		if v, ok := valueAtKey(raw, fmt.Sprintf("CurrentMap_%d", id)); ok { if n, ok := firstNumeric(v); ok { c.MapIndex = int(n); c.Map = fmt.Sprintf("Map #%d", int(n)) } else if s, ok := v.(string); ok { c.Map = s } }
		if v, ok := valueAtKey(raw, fmt.Sprintf("AFKtarget_%d", id)); ok { if s, ok := v.(string); ok { c.AFKTarget = s } }
		out = append(out, c)
	}
	// Some community exports include an explicit current character/player index.
	for _, key := range []string{"currentPlayer", "currentCharacter", "playerSelected", "CurrentPlayer"} {
		if v, ok := valueAtKey(raw, key); ok { if n, ok := firstNumeric(v); ok { for i := range out { if ids[i] == int(n) { out[i].Active = true } } }; break }
	}
	return out
}

func flattenKeyMatches(v any, signatures []string, depth int, path string, found *[]string, samples *[]any) {
	if depth < 0 { return }
	switch t := v.(type) {
	case map[string]any:
		for k, x := range t {
			p := k; if path != "" { p = path + "." + k }
			lk := strings.ToLower(k)
			matched := false
			for _, sig := range signatures { if strings.Contains(lk, strings.ToLower(sig)) { matched = true; break } }
			if matched { *found = append(*found, p); *samples = append(*samples, x) }
			flattenKeyMatches(x, signatures, depth-1, p, found, samples)
		}
	case []any:
		for i, x := range t { flattenKeyMatches(x, signatures, depth-1, fmt.Sprintf("%s[%d]", path, i), found, samples) }
	}
}

func progressHeuristic(samples []any) float64 {
	var total, nonzero float64
	var walk func(any, int)
	walk = func(v any, depth int) {
		if depth < 0 || total > 2500 { return }
		v = deepParseJSONish(v)
		switch t := v.(type) {
		case float64:
			total++; if t > 0 { nonzero++ }
		case bool:
			total++; if t { nonzero++ }
		case []any:
			for _, x := range t { walk(x, depth-1) }
		case map[string]any:
			for _, x := range t { walk(x, depth-1) }
		}
	}
	for _, s := range samples { walk(s, 3) }
	if total == 0 { return 0 }
	p := nonzero / total
	if p < 0 { return 0 }; if p > 1 { return 1 }; return p
}

func detectRawSystems(raw map[string]any) (map[string]SystemState, []string) {
	out := map[string]SystemState{}
	var detected []string
	for id, sigs := range rawSystemSignatures {
		var paths []string; var samples []any
		flattenKeyMatches(raw, sigs, 5, "", &paths, &samples)
		if len(paths) == 0 { continue }
		sort.Strings(paths)
		if len(paths) > 5 { paths = paths[:5] }
		p := progressHeuristic(samples)
		out[id] = SystemState{Progress:p, Ready:true, Evidence:"Raw save keys: "+strings.Join(paths, ", "), Confidence:.35, DetectedOnly:true}
		detected = append(detected, id)
	}
	sort.Strings(detected)
	return out, detected
}

func inferWorldFromSystems(systems map[string]SystemState) int {
	world := 1
	for id := range systems { if rule, ok := idleonSystems[id]; ok && rule.MinWorld > world { world = rule.MinWorld } }
	return world
}
