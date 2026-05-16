package app

type Tab int

const (
	TabOverview Tab = iota
	TabMessages
	TabChat
	TabCodex
	TabGemini
	TabMore
	TabCount
)

var tabNames = [TabCount]string{"Overview", "Messages", "Chat", "Codex", "Gemini", "More"}

func (t Tab) String() string {
	if int(t) < len(tabNames) {
		return tabNames[t]
	}
	return "Unknown"
}

func (t Tab) IsChannelType() bool {
	return t == TabMessages || t == TabChat || t == TabCodex || t == TabGemini
}
