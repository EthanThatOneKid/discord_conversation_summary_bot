package store

import (
	"github.com/diamondburned/arikawa/v3/gateway"
	"libdb.so/persist"
)

// Summary is a stored conversation summary.
type Summary = gateway.ConversationSummary

// Store is a persistent storage for conversation summaries.
type Store struct {
	db *persist.MustMap[string, Summary]
}

// Add adds a conversation summary to the store.
func (s *Store) Add(summary Summary) {
	s.db.Store(summary.ID.String(), summary)
}

// Get gets a conversation summary from the store.
func (s *Store) Get(id string) (Summary, bool) {
	summary, ok := s.db.Load(id)
	return summary, ok
}

// NewStore creates a new store.
func NewStore(db *persist.MustMap[string, Summary]) *Store {
	return &Store{
		db: db,
	}
}
