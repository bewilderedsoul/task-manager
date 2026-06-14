package tasks

import "testing"

func TestBuildOrderClause(t *testing.T) {
	cases := []struct {
		name     string
		sortBy   string
		order    string
		contains string
	}{
		{"default created_at desc", "created_at", "desc", "ORDER BY created_at DESC"},
		{"created_at asc", "created_at", "asc", "ORDER BY created_at ASC"},
		{"due date keeps nulls last", "due_date", "asc", "due_date ASC NULLS LAST"},
		{"priority is ranked not alphabetical", "priority", "desc", "CASE priority WHEN 'high' THEN 3"},
		{"unknown field falls back to created_at", "bogus", "asc", "ORDER BY created_at ASC"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := buildOrderClause(c.sortBy, c.order)
			if !contains(got, c.contains) {
				t.Errorf("buildOrderClause(%q,%q) = %q, want it to contain %q", c.sortBy, c.order, got, c.contains)
			}
		})
	}
}

func contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) && indexOf(haystack, needle) >= 0
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
