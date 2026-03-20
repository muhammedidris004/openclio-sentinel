package serving

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

// WorkspaceFileProvider serves tier-3 semantic memory from dataDir/memory.md.
type WorkspaceFileProvider struct {
	memoryFilePath string
}

// NewWorkspaceFileProvider creates a MemoryProvider that reads memory.md from dataDir.
func NewWorkspaceFileProvider(dataDir string) *WorkspaceFileProvider {
	return &WorkspaceFileProvider{
		memoryFilePath: filepath.Join(dataDir, "memory.md"),
	}
}

// GetMemories returns parsed memory lines from memory.md.
func (p *WorkspaceFileProvider) GetMemories(limit int) ([]string, error) {
	if p == nil || strings.TrimSpace(p.memoryFilePath) == "" {
		return nil, nil
	}

	data, err := os.ReadFile(p.memoryFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading memory file: %w", err)
	}

	return parseMemoryText(string(data), limit), nil
}

// GetMemoriesForQuery returns only the memory lines relevant to the current query.
func (p *WorkspaceFileProvider) GetMemoriesForQuery(query string, limit int) ([]string, error) {
	memories, err := p.GetMemories(0)
	if err != nil || len(memories) == 0 {
		return memories, err
	}
	return filterMemoriesForQuery(memories, query, limit), nil
}

// StaticProvider serves tier-3 semantic memory from an in-memory text blob.
type StaticProvider struct {
	memories []string
}

// NewStaticProvider parses memory text and returns a static MemoryProvider.
func NewStaticProvider(rawMemory string) *StaticProvider {
	return &StaticProvider{memories: parseMemoryText(rawMemory, 0)}
}

// GetMemories returns parsed memory lines from in-memory content.
func (p *StaticProvider) GetMemories(limit int) ([]string, error) {
	if p == nil || len(p.memories) == 0 {
		return nil, nil
	}
	if limit <= 0 || limit >= len(p.memories) {
		out := make([]string, len(p.memories))
		copy(out, p.memories)
		return out, nil
	}
	out := make([]string, limit)
	copy(out, p.memories[:limit])
	return out, nil
}

// GetMemoriesForQuery returns only the in-memory entries relevant to the current query.
func (p *StaticProvider) GetMemoriesForQuery(query string, limit int) ([]string, error) {
	if p == nil || len(p.memories) == 0 {
		return nil, nil
	}
	return filterMemoriesForQuery(p.memories, query, limit), nil
}

func parseMemoryText(raw string, limit int) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	lines := strings.Split(raw, "\n")
	memories := make([]string, 0, len(lines))
	inCodeBlock := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "```") {
			inCodeBlock = !inCodeBlock
			continue
		}
		if inCodeBlock {
			continue
		}
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		trimmed = trimListPrefix(trimmed)
		if trimmed == "" {
			continue
		}
		memories = append(memories, trimmed)
		if limit > 0 && len(memories) >= limit {
			return memories
		}
	}

	if len(memories) == 0 {
		memories = append(memories, raw)
	}
	return memories
}

func trimListPrefix(line string) string {
	for _, prefix := range []string{"- ", "* ", "+ "} {
		if strings.HasPrefix(line, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(line, prefix))
		}
	}

	// Handle numbered lists like "1. item".
	i := 0
	for i < len(line) && line[i] >= '0' && line[i] <= '9' {
		i++
	}
	if i > 0 && i+1 < len(line) && line[i] == '.' && line[i+1] == ' ' {
		return strings.TrimSpace(line[i+2:])
	}

	return line
}

type scoredMemory struct {
	idx    int
	memory string
	score  float64
}

func filterMemoriesForQuery(memories []string, query string, limit int) []string {
	if len(memories) == 0 {
		return nil
	}
	if limit <= 0 || limit > len(memories) {
		limit = len(memories)
	}

	terms := queryTerms(query)
	if shouldBypassQueryFiltering(query, terms) {
		out := make([]string, limit)
		copy(out, memories[:limit])
		return out
	}

	scored := make([]scoredMemory, 0, len(memories))
	for idx, memory := range memories {
		score := memoryRelevanceScore(memory, query, terms)
		if score <= 0 {
			continue
		}
		scored = append(scored, scoredMemory{
			idx:    idx,
			memory: memory,
			score:  score,
		})
	}
	if len(scored) == 0 {
		return nil
	}

	sort.SliceStable(scored, func(i, j int) bool {
		if scored[i].score == scored[j].score {
			return scored[i].idx < scored[j].idx
		}
		return scored[i].score > scored[j].score
	})

	if limit > len(scored) {
		limit = len(scored)
	}
	out := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		out = append(out, scored[i].memory)
	}
	return out
}

func memoryRelevanceScore(memory, query string, terms []string) float64 {
	normalizedMemory := normalizeText(memory)
	if normalizedMemory == "" {
		return 0
	}

	normalizedQuery := normalizeText(query)
	score := 0.0
	if normalizedQuery != "" && strings.Contains(normalizedMemory, normalizedQuery) {
		score += 1.0
	}

	matches := 0
	for _, term := range terms {
		if strings.Contains(normalizedMemory, term) {
			matches++
		}
	}
	if matches == 0 {
		return 0
	}

	ratio := float64(matches) / float64(len(terms))
	if score == 0 && ratio < 0.34 {
		return 0
	}

	return score + ratio
}

func queryTerms(query string) []string {
	rawTerms := strings.FieldsFunc(strings.ToLower(query), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	if len(rawTerms) == 0 {
		return nil
	}

	stopWords := map[string]struct{}{
		"a": {}, "an": {}, "and": {}, "are": {}, "as": {}, "at": {}, "be": {},
		"can": {}, "do": {}, "for": {}, "from": {}, "get": {}, "how": {}, "i": {},
		"in": {}, "is": {}, "it": {}, "let": {}, "me": {}, "my": {}, "of": {},
		"on": {}, "or": {}, "so": {}, "that": {}, "the": {}, "this": {}, "to": {},
		"us": {}, "we": {}, "what": {}, "with": {}, "would": {}, "you": {}, "your": {},
	}

	terms := make([]string, 0, len(rawTerms))
	seen := make(map[string]struct{}, len(rawTerms))
	for _, term := range rawTerms {
		if len(term) < 2 {
			continue
		}
		if _, skip := stopWords[term]; skip {
			continue
		}
		if _, exists := seen[term]; exists {
			continue
		}
		seen[term] = struct{}{}
		terms = append(terms, term)
	}
	return terms
}

func normalizeText(input string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(input))), " ")
}

func shouldBypassQueryFiltering(query string, terms []string) bool {
	if len(terms) == 0 {
		return true
	}
	normalized := normalizeText(query)
	switch normalized {
	case "hello", "hi", "hey", "hey there", "hello there":
		return true
	default:
		return false
	}
}
