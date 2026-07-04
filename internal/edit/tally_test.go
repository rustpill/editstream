package edit

import "testing"

func TestTallyFold(t *testing.T) {
	edits := []Edit{
		{Wiki: "enwiki", User: "Ada"},
		{Wiki: "enwiki", User: "Grace"},
		{Wiki: "commonswiki", User: "Ada"},
		{Wiki: "enwiki", User: "Ada"},
	}
	// create empty tally
	tally := NewTally()

	for _, e := range edits {
		tally.Apply(e)
	}

	if got := tally.CountForWiki("enwiki"); got != 3 {
		t.Errorf("enwiki count = %d, want 3", got)
	}

	if got := tally.CountForWiki("commonswiki"); got != 1 {
		t.Errorf("commonswiki count = %d, want 1", got)
	}
}

func TestTallyReplayDeterministic(t *testing.T) {
	edits := []Edit{
		{Wiki: "enwiki"},
		{Wiki: "dewiki"},
		{Wiki: "enwiki"},
	}

	// helper to build tally
	build := func() *Tally {
		tl := NewTally()
		for _, e := range edits {
			tl.Apply(e)
		}
		return tl
	}

	first := build()
	second := build()

	if first.CountForWiki("enwiki") != 2 {
		t.Fatalf("expected enwiki=2 after replay, got %d", first.CountForWiki("enwiki"))
	}

	for _, w := range []string{"enwiki", "dewiki"} {
		if first.CountForWiki(w) != second.CountForWiki(w) {
			t.Errorf("replay mismatch for %s: %d vs %d",
				w, first.CountForWiki(w), second.CountForWiki(w))
		}
	}

}
