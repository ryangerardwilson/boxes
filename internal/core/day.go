package core

import (
	"encoding/json"
	"time"
)

type DayState struct {
	Date       string   `json:"date"`
	CheckedIDs []string `json:"checked_ids"`
}

func NewDayState(date string) DayState {
	return DayState{Date: date, CheckedIDs: []string{}}
}

func Today(now time.Time) string {
	return now.Local().Format(time.DateOnly)
}

func (s DayState) CheckedSet() map[string]bool {
	checked := map[string]bool{}
	for _, id := range s.CheckedIDs {
		checked[id] = true
	}
	return checked
}

func (s DayState) IsChecked(id string) bool {
	return s.CheckedSet()[id]
}

func (s DayState) StatusFor(id string, config Config) CompletionStatus {
	return config.StatusFor(id, s.CheckedSet())
}

func (s DayState) WithChecked(id string, checked bool, config Config) DayState {
	checkedSet := s.CheckedSet()
	targetIDs := config.LeafIDsFor(id)
	if len(targetIDs) == 0 {
		return s
	}

	if checked {
		for _, targetID := range targetIDs {
			checkedSet[targetID] = true
		}
	} else {
		for _, targetID := range targetIDs {
			delete(checkedSet, targetID)
		}
	}

	next := DayState{Date: s.Date, CheckedIDs: []string{}}
	for _, leafID := range config.LeafIDs() {
		if checkedSet[leafID] {
			next.CheckedIDs = append(next.CheckedIDs, leafID)
		}
	}
	return next
}

func (s DayState) Toggle(id string, config Config) DayState {
	return s.WithChecked(id, s.StatusFor(id, config) != StatusChecked, config)
}

func (s DayState) Reset() DayState {
	return DayState{Date: s.Date, CheckedIDs: []string{}}
}

func (s DayState) CompletedCount(config Config) int {
	checked := s.CheckedSet()
	count := 0
	for _, leafID := range config.LeafIDs() {
		if checked[leafID] {
			count++
		}
	}
	return count
}

func ParseDayState(data []byte, date string) (DayState, error) {
	var state DayState
	if err := json.Unmarshal(data, &state); err != nil {
		return DayState{}, err
	}
	if state.Date == "" {
		state.Date = date
	}
	if state.CheckedIDs == nil {
		state.CheckedIDs = []string{}
	}
	return state, nil
}

func MarshalDayState(state DayState) ([]byte, error) {
	return json.MarshalIndent(state, "", "  ")
}
