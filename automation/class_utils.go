package main

import "strings"

func normalizeClass(value string) string {
	v := strings.ToLower(strings.TrimSpace(value))
	v = strings.NewReplacer("-"," ","_"," ","  "," ").Replace(v)
	switch v {
	case "bubo", "bubonic conjuror", "bubonic conjurer": return "bubo"
	case "arcane cultist", "ac": return "arcane cultist"
	case "dk", "divine knight": return "dk"
	case "es", "elemental sorcerer": return "es"
	case "bb", "blood berserker": return "bb"
	case "bm", "beast master": return "bm"
	case "sb", "siege breaker": return "sb"
	case "vman", "vw", "voidwalker", "void walker": return "vman"
	case "maestro": return "maestro"
	case "journeyman", "jman": return "journeyman"
	case "wizard": return "wizard"
	case "shaman": return "shaman"
	case "squire": return "squire"
	case "barbarian": return "barbarian"
	case "bowman": return "bowman"
	case "hunter": return "hunter"
	}
	return v
}

func classMatches(characterClass, preferred string) bool {
	if strings.TrimSpace(preferred)=="" { return true }
	return normalizeClass(characterClass)==normalizeClass(preferred)
}

func findBestCharacter(snapshot AccountSnapshot, preferredClass string) *CharacterSnapshot {
	var best *CharacterSnapshot
	for i := range snapshot.Characters {
		c := &snapshot.Characters[i]
		if preferredClass!="" && !classMatches(c.Class, preferredClass) { continue }
		if best==nil || c.Level>best.Level || (c.Active && !best.Active) { best=c }
	}
	if best!=nil { return best }
	for i:=range snapshot.Characters { if snapshot.Characters[i].Active { return &snapshot.Characters[i] } }
	for i:=range snapshot.Characters { c:=&snapshot.Characters[i]; if best==nil||c.Level>best.Level { best=c } }
	return best
}

func classRole(c CharacterSnapshot) string {
	switch normalizeClass(c.Class) {
	case "bubo": return "Alchemy / active mob farming"
	case "arcane cultist": return "Prisma Alchemy / advanced mage progression"
	case "dk": return "Construction / Death Note / crystal and boss farming"
	case "es": return "Resource farming / Worship / portal pushing"
	case "bb": return "Cooking / Zow-Chow / melee skilling"
	case "bm": return "Breeding / trapping support"
	case "sb": return "Catching / drop farming"
	case "vman": return "Account-wide skilling / Voidwalker milestones"
	default: return "General progression / account support"
	}
}
