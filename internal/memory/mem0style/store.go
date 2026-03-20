package mem0style

import (
	"sort"
	"strings"
	"sync"
	"time"
)

const defaultMinSalience = 0.70

// Fact is one consolidated memory unit in a Mem0-style fact store.
type Fact struct {
	ID         int64
	Claim      string
	Category   string
	Salience   float64
	ValidUntil *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// FactInput is used to add/update a fact in the store.
type FactInput struct {
	Claim      string
	Category   string
	Salience   float64
	ValidUntil *time.Time
}

// Store is an in-memory Mem0-style consolidated fact store.
//
// Design constraints:
// - consolidation uses exact normalized claim+category keys
// - conflicting claims are stored as separate facts
// - no epistemic contradiction resolution is performed
type Store struct {
	mu          sync.RWMutex
	nextID      int64
	minSalience float64
	factsByID   map[int64]*Fact
	keyToID     map[string]int64
}

// NewStore creates a Mem0-style fact store with conservative defaults.
func NewStore() *Store {
	return &Store{
		nextID:      1,
		minSalience: defaultMinSalience,
		factsByID:   make(map[int64]*Fact),
		keyToID:     make(map[string]int64),
	}
}

// SetMinSalience updates the minimum salience floor used by Upsert.
func (s *Store) SetMinSalience(v float64) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if v <= 0 {
		s.minSalience = defaultMinSalience
		return
	}
	s.minSalience = clampUnit(v)
}

// Upsert inserts or updates one consolidated fact by normalized claim+category.
func (s *Store) Upsert(input FactInput) *Fact {
	if s == nil {
		return nil
	}
	claim := normalizeText(input.Claim)
	if claim == "" {
		return nil
	}
	category := normalizeText(input.Category)
	if category == "" {
		category = "unknown"
	}
	salience := clampUnit(input.Salience)
	if salience < s.minSalience {
		salience = s.minSalience
	}

	key := factKey(claim, category)
	now := time.Now().UTC()

	s.mu.Lock()
	defer s.mu.Unlock()

	if id, ok := s.keyToID[key]; ok {
		existing := s.factsByID[id]
		if existing != nil {
			if salience > existing.Salience {
				existing.Salience = salience
			}
			if input.ValidUntil != nil {
				v := input.ValidUntil.UTC()
				existing.ValidUntil = &v
			}
			existing.UpdatedAt = now
			return cloneFact(existing)
		}
	}

	id := s.nextID
	s.nextID++
	fact := &Fact{
		ID:        id,
		Claim:     claim,
		Category:  category,
		Salience:  salience,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if input.ValidUntil != nil {
		v := input.ValidUntil.UTC()
		fact.ValidUntil = &v
	}
	s.factsByID[id] = fact
	s.keyToID[key] = id
	return cloneFact(fact)
}

// Get returns one fact by normalized claim+category key.
func (s *Store) Get(claim, category string) (*Fact, bool) {
	if s == nil {
		return nil, false
	}
	claim = normalizeText(claim)
	if claim == "" {
		return nil, false
	}
	category = normalizeText(category)
	if category == "" {
		category = "unknown"
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	id, ok := s.keyToID[factKey(claim, category)]
	if !ok {
		return nil, false
	}
	f := s.factsByID[id]
	if f == nil {
		return nil, false
	}
	return cloneFact(f), true
}

// List returns top facts ordered by salience and recency.
func (s *Store) List(limit int) []*Fact {
	if s == nil {
		return nil
	}
	if limit <= 0 {
		limit = 50
	}
	s.mu.RLock()
	all := make([]*Fact, 0, len(s.factsByID))
	for _, f := range s.factsByID {
		if f != nil {
			all = append(all, cloneFact(f))
		}
	}
	s.mu.RUnlock()

	sort.Slice(all, func(i, j int) bool {
		if all[i].Salience == all[j].Salience {
			return all[i].UpdatedAt.After(all[j].UpdatedAt)
		}
		return all[i].Salience > all[j].Salience
	})
	if len(all) > limit {
		all = all[:limit]
	}
	return all
}

// Search returns facts ranked by token overlap and salience.
func (s *Store) Search(query string, limit int) []*Fact {
	if s == nil {
		return nil
	}
	queryTokens := tokenize(normalizeText(query))
	if len(queryTokens) == 0 {
		return s.List(limit)
	}

	type scored struct {
		fact  *Fact
		score float64
	}
	s.mu.RLock()
	scoredFacts := make([]scored, 0, len(s.factsByID))
	for _, f := range s.factsByID {
		if f == nil {
			continue
		}
		tokens := tokenize(f.Claim)
		if len(tokens) == 0 {
			continue
		}
		overlap := overlapCount(queryTokens, tokens)
		if overlap == 0 {
			continue
		}
		score := float64(overlap) / float64(len(queryTokens))
		scoredFacts = append(scoredFacts, scored{
			fact:  cloneFact(f),
			score: score,
		})
	}
	s.mu.RUnlock()

	sort.Slice(scoredFacts, func(i, j int) bool {
		if scoredFacts[i].score == scoredFacts[j].score {
			if scoredFacts[i].fact.Salience == scoredFacts[j].fact.Salience {
				return scoredFacts[i].fact.UpdatedAt.After(scoredFacts[j].fact.UpdatedAt)
			}
			return scoredFacts[i].fact.Salience > scoredFacts[j].fact.Salience
		}
		return scoredFacts[i].score > scoredFacts[j].score
	})

	if limit <= 0 {
		limit = 10
	}
	if len(scoredFacts) > limit {
		scoredFacts = scoredFacts[:limit]
	}

	out := make([]*Fact, 0, len(scoredFacts))
	for _, item := range scoredFacts {
		out = append(out, item.fact)
	}
	return out
}

// CountExpired returns the number of facts that are expired at `now`.
func (s *Store) CountExpired(now time.Time) int {
	if s == nil {
		return 0
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	count := 0
	for _, f := range s.factsByID {
		if f == nil || f.ValidUntil == nil {
			continue
		}
		if f.ValidUntil.Before(now) {
			count++
		}
	}
	return count
}

func factKey(claim, category string) string {
	return claim + "|" + category
}

func cloneFact(f *Fact) *Fact {
	if f == nil {
		return nil
	}
	out := *f
	if f.ValidUntil != nil {
		v := *f.ValidUntil
		out.ValidUntil = &v
	}
	return &out
}

func normalizeText(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func clampUnit(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func tokenize(s string) map[string]struct{} {
	fields := strings.Fields(s)
	out := make(map[string]struct{}, len(fields))
	for _, f := range fields {
		f = strings.Trim(f, " .,;:!?()[]{}\"'`")
		if len(f) < 2 {
			continue
		}
		out[f] = struct{}{}
	}
	return out
}

func overlapCount(a, b map[string]struct{}) int {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	count := 0
	for token := range a {
		if _, ok := b[token]; ok {
			count++
		}
	}
	return count
}
