package main

import "time"

type GameState struct {
	Running     bool      `json:"running"`
	Foreground  bool      `json:"foreground"`
	WindowTitle string    `json:"windowTitle,omitempty"`
	X           int       `json:"x,omitempty"`
	Y           int       `json:"y,omitempty"`
	Width       int       `json:"width,omitempty"`
	Height      int       `json:"height,omitempty"`
	CheckedAt   time.Time `json:"checkedAt"`
}

type SourceDiagnostic struct {
	Source       string   `json:"source"`
	Schema       string   `json:"schema"`
	Loaded       bool     `json:"loaded"`
	Fresh        bool     `json:"fresh,omitempty"`
	CapturedAt   time.Time `json:"capturedAt,omitempty"`
	Characters   int      `json:"characters,omitempty"`
	Systems      int      `json:"systems,omitempty"`
	RawKeys      int      `json:"rawKeys,omitempty"`
	Warnings     []string `json:"warnings,omitempty"`
}

type SystemCoverage struct {
	ID         string `json:"id"`
	Label      string `json:"label"`
	World      int    `json:"world"`
	Status     string `json:"status"` // parsed, detected, missing, not-unlocked
	Confidence int    `json:"confidence"`
	Evidence   string `json:"evidence,omitempty"`
}

type AccountSnapshot struct {
	Source       string                 `json:"source"`
	Schema       string                 `json:"schema,omitempty"`
	AccountName  string                 `json:"accountName,omitempty"`
	World        int                    `json:"world,omitempty"`
	Characters   []CharacterSnapshot    `json:"characters,omitempty"`
	Systems      map[string]SystemState `json:"systems,omitempty"`
	Timers       []TimerState           `json:"timers,omitempty"`
	GreenStacks  []GreenStackState      `json:"greenStacks,omitempty"`
	Raw          map[string]any         `json:"raw,omitempty"`
	CapturedAt   time.Time              `json:"capturedAt"`
	Warnings     []string               `json:"warnings,omitempty"`
	DetectedKeys []string               `json:"detectedKeys,omitempty"`
}

type CharacterSnapshot struct {
	Name        string             `json:"name"`
	Class       string             `json:"class,omitempty"`
	ClassIndex  int                `json:"classIndex,omitempty"`
	Level       int                `json:"level,omitempty"`
	Active      bool               `json:"active,omitempty"`
	Map         string             `json:"map,omitempty"`
	MapIndex    int                `json:"mapIndex,omitempty"`
	AFKTarget   string             `json:"afkTarget,omitempty"`
	Skills      map[string]float64 `json:"skills,omitempty"`
	DamageScore float64            `json:"damageScore,omitempty"`
	Skilling    float64            `json:"skillingScore,omitempty"`
}

type SystemState struct {
	Progress       float64 `json:"progress"`
	Ready          bool    `json:"ready,omitempty"`
	Urgency        float64 `json:"urgency,omitempty"`
	AccountWide    float64 `json:"accountWide,omitempty"`
	UnlockValue    float64 `json:"unlockValue,omitempty"`
	Ease           float64 `json:"ease,omitempty"`
	Hours          float64 `json:"hours,omitempty"`
	Routine        string  `json:"routine,omitempty"`
	PreferredClass string  `json:"preferredClass,omitempty"`
	Current        float64 `json:"current,omitempty"`
	Target         float64 `json:"target,omitempty"`
	Note           string  `json:"note,omitempty"`
	Evidence       string  `json:"evidence,omitempty"`
	Confidence     float64 `json:"confidence,omitempty"`
	DetectedOnly   bool    `json:"detectedOnly,omitempty"`
}

type TimerState struct {
	Name      string    `json:"name"`
	Ready     bool      `json:"ready"`
	ReadyAt   time.Time `json:"readyAt,omitempty"`
	Routine   string    `json:"routine,omitempty"`
	Priority  float64   `json:"priority,omitempty"`
}

type GreenStackState struct {
	Name      string  `json:"name"`
	Count     float64 `json:"count"`
	Complete  bool    `json:"complete"`
	Map       string  `json:"map,omitempty"`
	Routine   string  `json:"routine,omitempty"`
	BestClass string  `json:"bestClass,omitempty"`
}

type Recommendation struct {
	ID               string   `json:"id"`
	Category         string   `json:"category"`
	Title            string   `json:"title"`
	Reason           string   `json:"reason"`
	Score            float64  `json:"score"`
	Confidence       float64  `json:"confidence"`
	System           string   `json:"system,omitempty"`
	Character        string   `json:"character,omitempty"`
	PreferredClass   string   `json:"preferredClass,omitempty"`
	Routine          string   `json:"routine,omitempty"`
	SwitchRoutine    string   `json:"switchRoutine,omitempty"`
	Action           string   `json:"action,omitempty"`
	Automatable      bool     `json:"automatable"`
	EstimatedMinutes int      `json:"estimatedMinutes,omitempty"`
	Dependencies     []string `json:"dependencies,omitempty"`
}

type CharacterPlan struct {
	Character string `json:"character"`
	Class     string `json:"class,omitempty"`
	Role      string `json:"role"`
	Reason    string `json:"reason"`
	Priority  int    `json:"priority"`
}

type Assessment struct {
	GeneratedAt       time.Time           `json:"generatedAt"`
	Game              GameState           `json:"game"`
	Source            string              `json:"source"`
	Schema            string              `json:"schema,omitempty"`
	AccountName       string              `json:"accountName,omitempty"`
	World             int                 `json:"world,omitempty"`
	Stage             string              `json:"stage,omitempty"`
	HealthScore       int                 `json:"healthScore"`
	CoveragePercent   int                 `json:"coveragePercent"`
	SnapshotAgeSec    int64               `json:"snapshotAgeSeconds,omitempty"`
	ActiveCharacter   *CharacterSnapshot  `json:"activeCharacter,omitempty"`
	Top               []Recommendation    `json:"top"`
	QuickWins         []Recommendation    `json:"quickWins"`
	TimersReady       []Recommendation    `json:"timersReady"`
	GreenStacks       []Recommendation    `json:"greenStacks"`
	CharacterPlan     []CharacterPlan     `json:"characterPlan,omitempty"`
	SystemCoverage    []SystemCoverage    `json:"systemCoverage,omitempty"`
	SourceDiagnostics []SourceDiagnostic  `json:"sourceDiagnostics,omitempty"`
	Warnings          []string            `json:"warnings,omitempty"`
}
