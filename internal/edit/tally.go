package edit

import "sort"

// Current edit counts per wiki, built by folds
// the edit log: state = fold(Apply, events)
type Tally struct {
	perWiki map[string]int
}

func NewTally() *Tally {
	return &Tally{perWiki: make(map[string]int)}
}

// 'fold' Edit into the tally, (increment counter)
func (t *Tally) Apply(e Edit) {
	t.perWiki[e.Wiki]++
}

// how many edits have been folded in for a wiki, (count)
func (t *Tally) CountForWiki(wiki string) int {
	return t.perWiki[wiki]
}

// WikiCount tuple struct
type WikiCount struct {
	Wiki  string
	Count int
}

func (t *Tally) TopWikis(n int) []WikiCount {
	// init empty with len(t.perWiki) capacity
	out := make([]WikiCount, 0, len(t.perWiki))
	for w, c := range t.perWiki {
		out = append(out, WikiCount{Wiki: w, Count: c})
	}
	// descending order
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Wiki < out[j].Wiki
	})
	// n elements only
	if n < len(out) {
		out = out[:n]
	}
	return out
}
