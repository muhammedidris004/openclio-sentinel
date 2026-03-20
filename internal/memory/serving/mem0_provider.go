package serving

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/openclio/openclio/internal/memory/mem0style"
)

const defaultMem0Category = "workspace_memory"

// Mem0WorkspaceConfig configures the Mem0-style workspace provider.
type Mem0WorkspaceConfig struct {
	DBPath      string
	MinSalience float64
}

// Mem0WorkspaceProvider serves tier-3 memories by consolidating memory.md
// into a local Mem0-style SQLite fact store.
type Mem0WorkspaceProvider struct {
	mu             sync.Mutex
	memoryFilePath string
	store          *mem0style.SQLiteStore
	lastSnapshot   string
}

// NewMem0WorkspaceProvider creates a Mem0-style memory provider.
func NewMem0WorkspaceProvider(dataDir string, cfg Mem0WorkspaceConfig) (*Mem0WorkspaceProvider, error) {
	dbPath := strings.TrimSpace(cfg.DBPath)
	if dbPath == "" {
		dbPath = filepath.Join(dataDir, "memory_mem0style.db")
	} else if !filepath.IsAbs(dbPath) {
		dbPath = filepath.Join(dataDir, dbPath)
	}

	store, err := mem0style.NewSQLiteStore(dbPath)
	if err != nil {
		return nil, fmt.Errorf("new mem0 sqlite store: %w", err)
	}
	if cfg.MinSalience > 0 {
		store.SetMinSalience(cfg.MinSalience)
	}

	return &Mem0WorkspaceProvider{
		memoryFilePath: filepath.Join(dataDir, "memory.md"),
		store:          store,
	}, nil
}

// Close closes the underlying Mem0-style store.
func (p *Mem0WorkspaceProvider) Close() error {
	if p == nil || p.store == nil {
		return nil
	}
	return p.store.Close()
}

// GetMemories returns consolidated memories from the Mem0-style store.
func (p *Mem0WorkspaceProvider) GetMemories(limit int) ([]string, error) {
	if p == nil || p.store == nil {
		return nil, nil
	}
	if err := p.sync(); err != nil {
		return nil, err
	}

	facts, err := p.store.List(limit)
	if err != nil {
		return nil, fmt.Errorf("list mem0 facts: %w", err)
	}
	if len(facts) == 0 {
		return nil, nil
	}
	out := make([]string, 0, len(facts))
	for _, fact := range facts {
		if fact == nil || strings.TrimSpace(fact.Claim) == "" {
			continue
		}
		out = append(out, fact.Claim)
	}
	return out, nil
}

// GetMemoriesForQuery returns only the consolidated memories relevant to the current query.
func (p *Mem0WorkspaceProvider) GetMemoriesForQuery(query string, limit int) ([]string, error) {
	memories, err := p.GetMemories(0)
	if err != nil || len(memories) == 0 {
		return memories, err
	}
	return filterMemoriesForQuery(memories, query, limit), nil
}

func (p *Mem0WorkspaceProvider) sync() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	data, err := os.ReadFile(p.memoryFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			if p.lastSnapshot == "" {
				return nil
			}
			if err := p.store.DeleteAll(); err != nil {
				return fmt.Errorf("clear mem0 store for missing memory file: %w", err)
			}
			p.lastSnapshot = ""
			return nil
		}
		return fmt.Errorf("read memory file: %w", err)
	}
	snapshot := string(data)
	if snapshot == p.lastSnapshot {
		return nil
	}
	memories := parseMemoryText(string(data), 0)

	if err := p.store.DeleteAll(); err != nil {
		return fmt.Errorf("reset mem0 store: %w", err)
	}
	for _, m := range memories {
		_, upsertErr := p.store.Upsert(mem0style.FactInput{
			Claim:    m,
			Category: defaultMem0Category,
			Salience: 0.90,
		})
		if upsertErr != nil {
			return fmt.Errorf("upsert mem0 fact: %w", upsertErr)
		}
	}
	p.lastSnapshot = snapshot
	return nil
}
