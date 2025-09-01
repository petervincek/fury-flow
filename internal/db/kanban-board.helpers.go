package db

import "fmt"

func (k KanbanBoard) String() string {
	return fmt.Sprintf("KanbanBoard{ID: %v, Name: %q}", k.BoardID, k.BoardName)
}
