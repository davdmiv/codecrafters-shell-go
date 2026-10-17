package commands

import (
	"fmt"
	"strconv"

	"github.com/codecrafters-io/shell-starter-go/app/history"
)

// HistoryCmd is the shell builtin `history`.
type HistoryCmd struct {
	store *history.History
}

func NewHistory(store *history.History) *HistoryCmd {
	return &HistoryCmd{store: store}
}

func (h *HistoryCmd) Name() string {
	return "history"
}

func (h *HistoryCmd) Execute(ctx *Context, args []string) error {
	limit := 0
	if len(args) > 0 {
		n, err := strconv.Atoi(args[0])
		if err != nil || n < 0 {
			return fmt.Errorf("history: %s: numeric argument required", args[0])
		}
		limit = n
	}

	entries := h.store.Last(limit)
	for _, e := range entries {
		if _, err := fmt.Fprintln(ctx.Stdout, history.FormatEntry(e)); err != nil {
			return err
		}
	}
	return nil
}
