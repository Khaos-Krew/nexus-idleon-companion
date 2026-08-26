package main

import "time"

type GameState struct {
	Running     bool      `json:"running"`
	Foreground  bool      `json:"foreground"`
	WindowTitle string    `json:"windowTitle,omitempty"`
	CheckedAt   time.Time `json:"checkedAt"`
}

type AccountSnapshot struct {
	Source       string                 `json:"source"`
	AccountName  string                 `json:"accountName,omitempty"`
	World        int                    `json:"world,omitempty"`
	Characters   []CharacterSnapshot    `json:"characters,omitempty"`
	Systems      map[string]SystemState `json:"systems,omitempty"`
	Timers       []TimerState           `json:"timers,omitempty"`
	GreenStacks  []GreenStackState      `json:"greenStacks,omitempty"`
	Raw          map[string]any         `json:"raw,omitempty"`
	CapturedAt   time.Time              `json:"capturedAt"`
}

type CharacterSnapshot struct {
	Name        string             `json:"name"`
	Class       string             `json:"class,omitempty"`
	Level       int                `json:"level,omitempty"`
	Active      bool               `json:"active,omitempty"`
	Map         string             `json:"map,omitempty"`
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
	Note           string  `json:"note,omitempty"`
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
	ID             string  `json:"id"`
	Category       string  `json:"category"`
	Title          string  `json:"title"`
	Reason         string  `json:"reason"`
	Score          float64 `json:"score"`
	System         string  `json:"system,omitempty"`
	Character      string  `json:"character,omitempty"`
	PreferredClass string  `json:"preferredClass,omitempty"`
	Routine        string  `json:"routine,omitempty"`
	Action         string  `json:"action,omitempty"`
	Automatable    bool    `json:"automatable"`
}

type Assessment struct {
	GeneratedAt    time.Time        `json:"generatedAt"`
	Game           GameState        `json:"game"`
	Source         string           `json:"source"`
	AccountName    string           `json:"accountName,omitempty"`
	World          int              `json:"world,omitempty"`
	HealthScore    int              `json:"healthScore"`
	ActiveCharacter *CharacterSnapshot `json:"activeCharacter,omitempty"`
	Top            []Recommendation `json:"top"`
	QuickWins      []Recommendation `json:"quickWins"`
	TimersReady    []Recommendation `json:"timersReady"`
	GreenStacks    []Recommendation `json:"greenStacks"`
}
