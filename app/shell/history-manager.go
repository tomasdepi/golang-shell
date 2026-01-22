package shell

import (
	"os"
	"strings"
)

type History interface {
	Prev() (string, bool)
	Next() (string, bool)
	Add(cmd string)
	ResetNav()
}

type HistoryManager struct {
	entries        []string
	cursor         int
	lastSavedIndex int
}

func NewHistory() *HistoryManager {

	history := &HistoryManager{
		entries: []string{},
		cursor:  0,
	}

	histFile := os.Getenv("HISTFILE")

	if histFile == "" {
		return history
	}

	history.LoadFromFile(histFile)

	return history
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

func (hm *HistoryManager) LoadFromFile(file string) error {

	fileContent, err := os.ReadFile(file)

	if err != nil {
		return err
	}

	newEntries := strings.Split(string(fileContent), "\n")

	for _, e := range newEntries {
		if e != "" {
			hm.entries = append(hm.entries, e)
		}
	}

	//hm.entries = append(hm.entries, newEntries...)
	hm.cursor = len(hm.entries)

	return nil
}

func (hm *HistoryManager) SaveToFile(file string) error {

	buff := strings.Join(append(hm.entries, ""), "\n")

	err := os.WriteFile(file, []byte(buff), 0644)

	if err != nil {
		return err
	}

	return nil
}

func (hm *HistoryManager) AppendToFile(file string) error {

	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	for _, entry := range hm.entries[hm.lastSavedIndex:] {
		f.WriteString(entry + "\n")
	}

	hm.lastSavedIndex = len(hm.entries)

	return nil
}
