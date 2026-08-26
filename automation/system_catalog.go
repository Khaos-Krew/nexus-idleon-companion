package main

type SystemRule struct {
	Label          string
	MinWorld       int
	AccountWide    float64
	UnlockValue    float64
	Ease           float64
	Hours          float64
	PreferredClass string
	Action         string
}

var idleonSystems = map[string]SystemRule{
	"stamps": {"Stamps", 1, 5, 4, 4, .35, "", "Buy efficient stamp levels, prioritizing cheap account-wide gains."},
	"forge": {"Forge", 1, 3, 3, 4, .25, "", "Keep forge slots and ore conversion aligned to current crafting needs."},
	"anvil": {"Anvil", 1, 4, 4, 4, .25, "", "Spend points, improve production speed/capacity, and assign materials needed by account goals."},
	"alchemy": {"Alchemy", 2, 5, 5, 3, 1, "Bubo", "Push bubbles, vials, liquids and cauldrons; use Bubo when active alchemy progression is the best return."},
	"postoffice": {"Post Office", 2, 4, 3, 4, .25, "", "Spend available boxes on current character/account priorities."},
	"obols": {"Obols", 2, 3, 3, 4, .25, "", "Fill empty family/player slots and improve obols for the active objective."},
	"refinery": {"Refinery", 3, 5, 5, 3, .5, "", "Protect salt ratios and prevent higher-tier production from starving base salts."},
	"construction": {"Construction", 3, 5, 5, 3, 1.5, "DK", "Prioritize build rate, key buildings and tower levels that unlock account progression."},
	"printer": {"3D Printer", 3, 5, 5, 3, .75, "", "Resample valuable materials on the best character and keep printer output aligned to bottlenecks."},
	"worship": {"Worship", 3, 4, 4, 4, .25, "ES", "Spend charge, improve skulls and push tower defense breakpoints for souls/prayers."},
	"trapping": {"Trapping", 3, 4, 4, 4, .25, "", "Refresh traps with better boxes and targets; avoid capped or stale trap cycles."},
	"shrines": {"Shrines", 3, 3, 3, 4, .15, "", "Place and level shrines where the active account push benefits most."},
	"cooking": {"Cooking", 4, 5, 5, 3, 1, "BB", "Unlock meals and push efficient meal levels before expensive single-meal grinds."},
	"breeding": {"Breeding", 4, 4, 5, 2, 2, "BM", "Push pet power, territory and arena breakpoints that unlock cooking multipliers."},
	"lab": {"Laboratory", 4, 5, 5, 3, .75, "", "Maintain important lab nodes while minimizing unnecessary character lock-in."},
	"rift": {"Rift", 4, 5, 5, 2, 2, "", "Push Rift breakpoints when the next unlock materially improves account progression."},
	"divinity": {"Divinity", 5, 5, 5, 3, 1, "", "Unlock and assign gods that improve AFK, lab and skilling progression."},
	"sailing": {"Sailing", 5, 5, 5, 3, .5, "", "Collect loot, improve boats and target high-value artifacts."},
	"gaming": {"Gaming", 5, 4, 4, 4, .25, "", "Collect gains and buy compounding gaming upgrades before capping resources."},
	"companions": {"Companions", 5, 4, 4, 4, .15, "", "Review available companion bonuses and account-wide synergy."},
	"sneaking": {"Sneaking", 6, 4, 4, 3, .5, "", "Keep all available characters progressing with suitable items and floor targets."},
	"farming": {"Farming", 6, 4, 4, 3, .5, "", "Harvest on schedule and prioritize crop unlocks/upgrades that compound future gains."},
	"summoning": {"Summoning", 6, 4, 5, 2, 1.5, "", "Push match breakpoints and slime upgrades where a win unlocks a major multiplier."},
	"owl": {"Owl", 1, 3, 3, 3, .5, "", "Claim and reinvest when the next Owl breakpoint is efficient."},
	"roo": {"Roo", 2, 3, 3, 3, .5, "", "Claim and reinvest Roo progress when the next breakpoint is efficient."},
	"cards": {"Cards", 1, 3, 3, 4, .25, "", "Use card sets that match the current activity rather than generic damage everywhere."},
	"constellations": {"Star Signs", 1, 3, 3, 4, .2, "", "Unlock and equip star signs that support the current objective."},
	"gear": {"Gear & Tools", 1, 3, 4, 3, 1, "", "Upgrade weapons, armor and tools where they unlock meaningful breakpoints."},
	"talents": {"Talents", 1, 3, 3, 5, .1, "", "Fix active/skilling presets and max talents that directly multiply the current goal."},
	"statues": {"Statues", 1, 4, 3, 3, .5, "", "Deposit and upgrade statues with strong account-wide returns."},
	"dungeons": {"Dungeons", 1, 4, 4, 2, 1, "", "Use weekly dungeon opportunities and spend credits/flurbos on high-value account upgrades."},
	"bosses": {"Bosses", 1, 3, 4, 3, .5, "DK", "Clear available boss tiers and cards/keys when they unlock meaningful upgrades."},
	"greenstacks": {"Green Stacks", 1, 5, 4, 3, 1, "ES", "Finish efficient missing green stacks and move permanently completed stacks out of the queue."},
	"deathnote": {"Death Note", 3, 5, 4, 2, 2, "DK", "Push efficient monster-family kill-count breakpoints for account-wide damage."},
	"tome": {"Tome", 4, 4, 4, 2, 1.5, "", "Target achievable Tome milestones that produce broad account benefits."},
}
