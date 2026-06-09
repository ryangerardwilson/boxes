package core

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode"
)

type Config struct {
	Items []Item `json:"items"`
}

type Item struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	Children []Item `json:"items,omitempty"`
}

type FlatItem struct {
	ID          string
	Label       string
	Depth       int
	HasChildren bool
}

type CompletionStatus int

const (
	StatusUnchecked CompletionStatus = iota
	StatusPartial
	StatusChecked
)

func DefaultConfig() Config {
	return Config{Items: []Item{}}
}

func StarterConfigText() string {
	return strings.TrimSpace(`# One box per line.
# Indent with two spaces for nested boxes.
#
# Move
# Work
#   Email
#   Deep work
# Plan tomorrow`) + "\n"
}

func ParseConfig(data []byte) (Config, error) {
	if looksLikeJSON(data) {
		return ParseJSONConfig(data)
	}
	return ParseOutlineConfig(string(data))
}

func ParseOutlineConfig(text string) (Config, error) {
	type pendingItem struct {
		depth int
		item  Item
	}

	var roots []Item
	stack := []pendingItem{}
	seen := map[string]struct{}{}

	for lineNumber, rawLine := range strings.Split(text, "\n") {
		if strings.TrimSpace(rawLine) == "" {
			continue
		}
		if strings.HasPrefix(strings.TrimSpace(rawLine), "#") {
			continue
		}

		depth, label, err := parseOutlineLine(rawLine)
		if err != nil {
			return Config{}, fmt.Errorf("line %d: %w", lineNumber+1, err)
		}
		if label == "" {
			continue
		}

		if depth > len(stack) {
			return Config{}, fmt.Errorf("line %d: indentation jumps more than one level", lineNumber+1)
		}
		stack = stack[:depth]

		pathParts := make([]string, 0, len(stack)+1)
		for _, parent := range stack {
			pathParts = append(pathParts, parent.item.ID)
		}
		pathParts = append(pathParts, normalizeItemID(label))
		id := strings.Join(pathParts, "/")
		if id == "" {
			return Config{}, fmt.Errorf("line %d: item label must contain a letter or number", lineNumber+1)
		}
		if _, ok := seen[id]; ok {
			return Config{}, fmt.Errorf("line %d: duplicate item id %q", lineNumber+1, id)
		}
		seen[id] = struct{}{}

		item := Item{ID: id, Label: label}
		if depth == 0 {
			roots = append(roots, item)
			stack = append(stack, pendingItem{depth: depth, item: item})
			continue
		}

		parent := &roots
		for _, ancestor := range stack {
			parentItem := findItem(parent, ancestor.item.ID)
			if parentItem == nil {
				return Config{}, fmt.Errorf("line %d: internal outline parent lookup failed", lineNumber+1)
			}
			parent = &parentItem.Children
		}
		*parent = append(*parent, item)
		stack = append(stack, pendingItem{depth: depth, item: item})
	}

	config := Config{Items: roots}
	if err := config.Validate(); err != nil {
		return Config{}, err
	}
	return config, nil
}

func ParseJSONConfig(data []byte) (Config, error) {
	var legacy struct {
		Items []legacyItem `json:"items"`
	}
	if err := json.Unmarshal(data, &legacy); err != nil {
		return Config{}, err
	}
	items := make([]Item, 0, len(legacy.Items))
	for _, item := range legacy.Items {
		items = append(items, item.toItem(nil))
	}
	config := Config{Items: items}
	if err := config.Validate(); err != nil {
		return Config{}, err
	}
	return config, nil
}

type legacyItem struct {
	ID       string
	Label    string
	Children []legacyItem
}

func (i *legacyItem) UnmarshalJSON(data []byte) error {
	var label string
	if err := json.Unmarshal(data, &label); err == nil {
		i.Label = strings.TrimSpace(label)
		i.ID = normalizeItemID(i.Label)
		return nil
	}

	var item struct {
		ID       string       `json:"id"`
		Label    string       `json:"label"`
		Items    []legacyItem `json:"items"`
		Children []legacyItem `json:"children"`
	}
	if err := json.Unmarshal(data, &item); err != nil {
		return err
	}
	i.Label = strings.TrimSpace(item.Label)
	i.ID = strings.TrimSpace(item.ID)
	if i.ID == "" {
		i.ID = normalizeItemID(i.Label)
	}
	if len(item.Items) > 0 {
		i.Children = item.Items
	} else {
		i.Children = item.Children
	}
	return nil
}

func (i legacyItem) toItem(parentPath []string) Item {
	idPart := strings.TrimSpace(i.ID)
	if idPart == "" {
		idPart = normalizeItemID(i.Label)
	}
	id := strings.Join(append(append([]string{}, parentPath...), idPart), "/")
	item := Item{ID: id, Label: strings.TrimSpace(i.Label)}
	for _, child := range i.Children {
		item.Children = append(item.Children, child.toItem(append(parentPath, idPart)))
	}
	return item
}

func MarshalConfig(config Config) ([]byte, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	return []byte(MarshalOutlineConfig(config)), nil
}

func MarshalOutlineConfig(config Config) string {
	var lines []string
	var walk func(items []Item, depth int)
	walk = func(items []Item, depth int) {
		for _, item := range items {
			lines = append(lines, strings.Repeat("  ", depth)+item.Label)
			walk(item.Children, depth+1)
		}
	}
	walk(config.Items, 0)
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}

func (c Config) Validate() error {
	seen := map[string]struct{}{}
	var walk func(items []Item) error
	walk = func(items []Item) error {
		for _, item := range items {
			id := strings.TrimSpace(item.ID)
			label := strings.TrimSpace(item.Label)
			if id == "" {
				return fmt.Errorf("item %q has empty id", label)
			}
			if label == "" {
				return fmt.Errorf("item %q has empty label", id)
			}
			if strings.ContainsAny(id, " \t\r\n") {
				return fmt.Errorf("item id %q must not contain whitespace", id)
			}
			if _, ok := seen[id]; ok {
				return fmt.Errorf("duplicate item id %q", id)
			}
			seen[id] = struct{}{}
			if err := walk(item.Children); err != nil {
				return err
			}
		}
		return nil
	}
	return walk(c.Items)
}

func (c Config) Flatten() []FlatItem {
	var flat []FlatItem
	var walk func(items []Item, depth int)
	walk = func(items []Item, depth int) {
		for _, item := range items {
			flat = append(flat, FlatItem{
				ID:          item.ID,
				Label:       item.Label,
				Depth:       depth,
				HasChildren: len(item.Children) > 0,
			})
			walk(item.Children, depth+1)
		}
	}
	walk(c.Items, 0)
	return flat
}

func (c Config) ItemByID(id string) (Item, bool) {
	for _, item := range c.Items {
		if found, ok := itemByID(item, id); ok {
			return found, true
		}
	}
	return Item{}, false
}

func (c Config) LeafIDsFor(id string) []string {
	item, ok := c.ItemByID(id)
	if !ok {
		return []string{}
	}
	return item.LeafIDs()
}

func (c Config) LeafIDs() []string {
	var ids []string
	for _, item := range c.Items {
		ids = append(ids, item.LeafIDs()...)
	}
	return ids
}

func (c Config) LeafCount() int {
	return len(c.LeafIDs())
}

func (c Config) StatusFor(id string, checkedSet map[string]bool) CompletionStatus {
	leafIDs := c.LeafIDsFor(id)
	if len(leafIDs) == 0 {
		return StatusUnchecked
	}

	checked := 0
	for _, leafID := range leafIDs {
		if checkedSet[leafID] {
			checked++
		}
	}
	if checked == 0 {
		return StatusUnchecked
	}
	if checked == len(leafIDs) {
		return StatusChecked
	}
	return StatusPartial
}

func (i Item) LeafIDs() []string {
	if len(i.Children) == 0 {
		return []string{i.ID}
	}
	var ids []string
	for _, child := range i.Children {
		ids = append(ids, child.LeafIDs()...)
	}
	return ids
}

func itemByID(item Item, id string) (Item, bool) {
	if item.ID == id {
		return item, true
	}
	for _, child := range item.Children {
		if found, ok := itemByID(child, id); ok {
			return found, true
		}
	}
	return Item{}, false
}

func findItem(items *[]Item, id string) *Item {
	for index := range *items {
		if (*items)[index].ID == id {
			return &(*items)[index]
		}
	}
	return nil
}

func parseOutlineLine(rawLine string) (int, string, error) {
	spaceCount := 0
	for _, r := range rawLine {
		switch r {
		case ' ':
			spaceCount++
		case '\t':
			spaceCount += 2
		default:
			goto done
		}
	}
done:
	if spaceCount%2 != 0 {
		return 0, "", fmt.Errorf("indent with two spaces per level")
	}
	label := strings.TrimSpace(rawLine)
	label = stripBullet(label)
	return spaceCount / 2, strings.TrimSpace(label), nil
}

func stripBullet(label string) string {
	for _, prefix := range []string{"- ", "* ", "+ "} {
		if strings.HasPrefix(label, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(label, prefix))
		}
	}
	if strings.HasPrefix(label, "[ ] ") {
		return strings.TrimSpace(strings.TrimPrefix(label, "[ ] "))
	}
	if strings.HasPrefix(label, "[x] ") || strings.HasPrefix(label, "[X] ") {
		return strings.TrimSpace(label[4:])
	}
	return label
}

func looksLikeJSON(data []byte) bool {
	for _, b := range data {
		switch b {
		case ' ', '\t', '\n', '\r':
			continue
		case '{', '[':
			return true
		default:
			return false
		}
	}
	return false
}

func normalizeItemID(label string) string {
	var b strings.Builder
	lastWasSeparator := false
	for _, r := range strings.ToLower(strings.TrimSpace(label)) {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
			lastWasSeparator = false
		case r == '-' || r == '_':
			if b.Len() > 0 && !lastWasSeparator {
				b.WriteRune(r)
				lastWasSeparator = true
			}
		default:
			if b.Len() > 0 && !lastWasSeparator {
				b.WriteRune('-')
				lastWasSeparator = true
			}
		}
	}
	return strings.Trim(b.String(), "-_")
}
