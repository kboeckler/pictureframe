package event

type Event struct {
	Summary string `json:"summary"`
	Start   string `json:"start"`
	Type    string `json:"type"`
}
