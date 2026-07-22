package ledger

import (
	"fmt"
	"sort"
)

// Entry records a signed amount for one account.
type Entry struct {
	Account string
	Amount  int64
}

func Summarize(entries []Entry) ([]string, error) {
	totals := make(map[string]int64)
	for _, entry := range entries {
		if entry.Account == "" {
			return nil, fmt.Errorf("entry has no account")
		}
		totals[entry.Account] += entry.Amount
	}
	accounts := make([]string, 0, len(totals))
	for account := range totals {
		accounts = append(accounts, account)
	}
	sort.Strings(accounts)
	for index, account := range accounts {
		accounts[index] = fmt.Sprintf("%s=%d", account, totals[account])
	}
	return accounts, nil
}
