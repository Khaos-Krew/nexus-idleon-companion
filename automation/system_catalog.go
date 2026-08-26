package main

type SystemRule struct {
	Label          string
	MinWorld       int
	AccountWide    float64
	UnlockValue    float64
	Ease           float64
	Hours          float64
	PreferredClass string
	Routine        string
	Dependencies   []string
	Action         string
}

var idleonSystems = map[string]SystemRule{
	"worldpush": {"World Progression",1,5,5,2,2,"ES","world-push",nil,"Push the next portal/boss only after the account-wide blockers above it are addressed."},
	"stamps": {"Stamps",1,5,4,4,.35,"","stamps",nil,"Buy efficient stamp levels, prioritizing cheap account-wide gains."},
	"vault": {"Vault",1,4,4,3,.5,"","vault",nil,"Claim and progress Vault unlocks that multiply account-wide systems."},
	"forge": {"Forge",1,3,3,4,.25,"","forge",nil,"Keep forge slots and ore conversion aligned to current crafting needs."},
	"anvil": {"Anvil / Smithing",1,4,4,4,.25,"","anvil",nil,"Spend points, improve speed/capacity, and assign materials needed by current goals."},
	"statues": {"Statues",1,4,3,3,.5,"","statues",nil,"Deposit and upgrade statues with strong account-wide returns."},
	"cards": {"Cards",1,3,3,4,.25,"","cards",nil,"Use card sets that match the current activity rather than generic damage everywhere."},
	"constellations": {"Star Signs",1,3,3,4,.2,"","constellations",nil,"Unlock and equip star signs that support the current objective."},
	"talents": {"Talents",1,3,3,5,.1,"","talents",nil,"Fix active/skilling presets and max talents that directly multiply the current goal."},
	"gear": {"Gear & Tools",1,3,4,3,1,"","gear",[]string{"stamps","alchemy"},"Upgrade weapons, armor and tools only where they unlock meaningful breakpoints."},
	"quests": {"Quests",1,3,4,3,.5,"","quests",nil,"Clear quests that unlock systems, maps, stamps, recipes, keys or account-wide features."},
	"tasks": {"Tasks & Merits",1,4,4,3,.5,"","tasks",nil,"Finish efficient task-board objectives and spend merit points on broad progression."},
	"dungeons": {"Party Dungeons",1,4,4,2,1,"","dungeons",nil,"Use weekly dungeon opportunities and spend credits/flurbos on high-value account upgrades."},
	"alchemy": {"Alchemy",2,5,5,3,1,"Bubo","alchemy",nil,"Push bubbles, vials, liquids and cauldrons; use Bubo when active alchemy is the best return."},
	"prisma": {"Prisma Bubbles",2,5,5,2,1.5,"Arcane Cultist","prisma",[]string{"alchemy"},"Advance Prisma Bubble progression when its account-wide multipliers beat cheaper normal Alchemy gains."},
	"postoffice": {"Post Office",2,4,3,4,.25,"","postoffice",nil,"Spend available boxes on current character/account priorities."},
	"obols": {"Obols",2,3,3,4,.25,"","obols",nil,"Fill empty family/player slots and improve obols for the active objective."},
	"refinery": {"Refinery",3,5,5,3,.5,"","refinery",nil,"Protect salt ratios and prevent higher-tier production from starving base salts."},
	"construction": {"Construction",3,5,5,3,1.5,"DK","construction",[]string{"refinery"},"Prioritize build rate, key buildings and tower levels that unlock account progression."},
	"printer": {"3D Printer",3,5,5,3,.75,"","printer",[]string{"gear","talents"},"Resample valuable materials on the best character and align printer output to bottlenecks."},
	"worship": {"Worship",3,4,4,4,.25,"ES","worship",[]string{"construction"},"Spend charge, improve skulls and push tower-defense breakpoints for souls/prayers."},
	"trapping": {"Trapping",3,4,4,4,.25,"","trapping",nil,"Refresh traps with better boxes and targets; avoid capped or stale trap cycles."},
	"shrines": {"Shrines",3,3,3,4,.15,"","shrines",[]string{"construction"},"Place and level shrines where the active account push benefits most."},
	"deathnote": {"Death Note",3,5,4,2,2,"DK","deathnote",nil,"Push efficient monster-family kill-count breakpoints for account-wide damage."},
	"cooking": {"Cooking",4,5,5,3,1,"BB","cooking",[]string{"breeding"},"Unlock meals and push efficient meal levels before expensive single-meal grinds."},
	"breeding": {"Breeding",4,4,5,2,2,"BM","breeding",nil,"Push pet power, territory and arena breakpoints that unlock cooking multipliers."},
	"lab": {"Laboratory",4,5,5,3,.75,"","lab",nil,"Maintain important lab nodes while minimizing unnecessary character lock-in."},
	"rift": {"Rift",4,5,5,2,2,"","rift",nil,"Push Rift breakpoints when the next unlock materially improves account progression."},
	"tome": {"Tome",4,4,4,2,1.5,"","tome",nil,"Target achievable Tome milestones that produce broad account benefits."},
	"killroy": {"Killroy",4,3,3,4,.25,"","killroy",nil,"Use available Killroy runs and spend skulls on account-relevant upgrades."},
	"divinity": {"Divinity",5,5,5,3,1,"","divinity",[]string{"lab"},"Unlock and assign gods that improve AFK, lab and skilling progression."},
	"sailing": {"Sailing",5,5,5,3,.5,"","sailing",nil,"Collect loot, improve boats and target high-value artifacts."},
	"gaming": {"Gaming",5,4,4,4,.25,"","gaming",nil,"Collect gains and buy compounding gaming upgrades before capping resources."},
	"companions": {"Companions",5,4,4,4,.15,"","companions",nil,"Review available companion bonuses and account-wide synergy."},
	"hole": {"The Hole",5,5,5,3,1,"","hole",nil,"Progress Hole villagers/features where the next unlock multiplies multiple older systems."},
	"slab": {"Slab",5,4,4,3,.75,"","slab",nil,"Fill efficient missing Slab entries and prioritize entries tied to broad account bonuses."},
	"sneaking": {"Sneaking",6,4,4,3,.5,"","sneaking",nil,"Keep all available characters progressing with suitable items and floor targets."},
	"farming": {"Farming",6,4,4,3,.5,"","farming",nil,"Harvest on schedule and prioritize crop unlocks/upgrades that compound future gains."},
	"summoning": {"Summoning",6,4,5,2,1.5,"","summoning",nil,"Push match breakpoints and slime upgrades where a win unlocks a major multiplier."},
	"beanstalk": {"Beanstalk",6,4,4,2,1,"","beanstalk",nil,"Push affordable Beanstalk food thresholds that create permanent account-wide bonuses."},
	"masterclasses": {"Master Classes",6,5,5,2,2,"","masterclasses",nil,"Advance unlocked Master Class progression when it opens major account-wide mechanics."},
	"research": {"Research",7,5,5,3,.75,"","research",nil,"Unlock observations and spend Research Points on the highest-impact account-wide nodes."},
	"spelunking": {"Spelunking",7,5,5,3,1,"","spelunking",nil,"Push tunnels, amber upgrades, page collection and Lore bosses to unlock W7 progression."},
	"coralreef": {"Coral Reef",7,4,4,3,.5,"","coralreef",[]string{"spelunking"},"Recover fishies and upgrade reef bonuses when they unlock meaningful account-wide gains."},
	"sushistation": {"Sushi Station",7,4,4,3,.5,"","sushistation",[]string{"research"},"Merge sushi efficiently, avoid capped fuel/resources and buy compounding Bucks upgrades."},
	"greenstacks": {"Green Stacks",1,5,4,3,1,"ES","greenstacks",nil,"Finish efficient missing green stacks and permanently remove completed 10M stacks from the queue."},
	"bosses": {"Bosses",1,3,4,3,.5,"DK","bosses",nil,"Clear available boss tiers and cards/keys when they unlock meaningful upgrades."},
	"minibosses": {"Minibosses",2,3,3,4,.25,"","minibosses",nil,"Avoid wasting capped miniboss spawns and collect efficient account resources/cards."},
	"colosseum": {"Colosseum",1,2,3,3,.5,"","colosseum",nil,"Use tickets when rewards or account unlocks justify the run."},
	"weeklyboss": {"Weekly Boss",4,3,4,3,.5,"","weeklyboss",nil,"Use available weekly boss attempts and prioritize permanent account rewards."},
	"vman": {"Voidwalker",4,5,5,2,2,"VMan","vman",nil,"Push Voidwalker speedrun/skilling milestones when their account-wide bonuses are reachable."},
	"owl": {"Owl",1,3,3,3,.5,"","owl",nil,"Claim and reinvest when the next Owl breakpoint is efficient."},
	"roo": {"Roo",2,3,3,3,.5,"","roo",nil,"Claim and reinvest Roo progress when the next breakpoint is efficient."},
}
