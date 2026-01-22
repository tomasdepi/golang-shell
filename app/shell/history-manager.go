package shell

type History interface {
	Prev() (string, bool)
	Next() (string, bool)
	Add(cmd string)
	ResetNav()
}

type HistoryManager struct {
	entries []string
	cursor  int
}

func NewHistory() *HistoryManager {
	return &HistoryManager{
		entries: []string{},
		cursor:  0,
	}
}

func (hm *HistoryManager) Add(cmd string) {
	if cmd == "" {
		return
	}
	hm.entries = append(hm.entries, cmd)
	hm.cursor = len(hm.entries)
}

func (hm *HistoryManager) Prev() (string, bool) {
	if hm.cursor == 0 {
		return "", false
	}
	hm.cursor--
	return hm.entries[hm.cursor], true
}

func (hm *HistoryManager) Next() (string, bool) {
	if hm.cursor >= len(hm.entries)-1 {
		hm.cursor = len(hm.entries)
		return "", false
	}
	hm.cursor++
	return hm.entries[hm.cursor], true
}

func (hm *HistoryManager) ResetNav() {
	hm.cursor = len(hm.entries)
}

func (hm *HistoryManager) ReadAll() []string {
	return hm.entries
}

func (hm *HistoryManager) ReadLastN(n int) []string {
	return hm.entries[len(hm.entries)-n:]
}

func (hm *HistoryManager) GetHistoryLen() int {
	return len(hm.entries)
}
