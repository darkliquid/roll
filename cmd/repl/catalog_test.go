package main

import (
	"testing"
)

func TestDefaultCatalogContainsSupportedComponents(t *testing.T) {
	cat := DefaultCatalog()
	if len(cat) == 0 {
		t.Fatal("expected non-empty default catalog")
	}

	// Verify required syntax tokens are present in catalog
	requiredSyntax := []string{
		"d6", "d%", "dF", // standard dice
		"!", "!!", "!p", // exploding
		"kh", "kl", "dh", "dl", // keep/drop
		"r", "ro", // reroll
		">", "<", "=", "f", // target / success / failure
		"s", "sd", // sorting
		"{}", // grouping
	}

	foundSyntax := make(map[string]bool)
	for _, comp := range cat {
		if comp.Name == "" {
			t.Errorf("component has empty Name: %+v", comp)
		}
		if comp.Syntax == "" {
			t.Errorf("component has empty Syntax: %+v", comp)
		}
		if comp.Template == "" {
			t.Errorf("component has empty Template: %+v", comp)
		}
		if comp.Category == "" {
			t.Errorf("component has empty Category: %+v", comp)
		}
		if comp.Description == "" {
			t.Errorf("component has empty Description: %+v", comp)
		}
		if comp.Example == "" {
			t.Errorf("component has empty Example: %+v", comp)
		}
		foundSyntax[comp.Syntax] = true
	}

	for _, syntax := range requiredSyntax {
		if !foundSyntax[syntax] {
			t.Errorf("expected catalog to contain component with syntax %q", syntax)
		}
	}
}

func TestFilterComponentsByPrefix(t *testing.T) {
	cat := DefaultCatalog()

	t.Run("ExtractTargetToken", func(t *testing.T) {
		tests := []struct {
			name     string
			input    string
			cursor   int
			expected string
		}{
			{"empty input", "", 0, ""},
			{"cursor at 0", "4d6", 0, ""},
			{"token k at end", "4d6k", 4, "k"},
			{"token kh at end", "4d6kh", 5, "kh"},
			{"token ! at end", "4d6!", 4, "!"},
			{"token d in middle", "4d6", 2, "d"},
			{"space prefix", "4d6 + 2d", 8, "d"},
			{"cursor beyond end clamped", "4d6k", 10, "k"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := ExtractTargetToken(tt.input, tt.cursor)
				if got != tt.expected {
					t.Errorf("ExtractTargetToken(%q, %d) = %q; want %q", tt.input, tt.cursor, got, tt.expected)
				}
			})
		}
	})

	t.Run("FilterComponents", func(t *testing.T) {
		tests := []struct {
			name           string
			input          string
			cursor         int
			expectedSyntax []string
		}{
			{
				name:           "prefix k filters kh and kl",
				input:          "4d6k",
				cursor:         4,
				expectedSyntax: []string{"kh", "kl"},
			},
			{
				name:           "prefix ! filters !, !!, !p",
				input:          "4d6!",
				cursor:         4,
				expectedSyntax: []string{"!", "!!", "!p"},
			},
			{
				name:           "prefix r filters r and ro",
				input:          "4d6r",
				cursor:         4,
				expectedSyntax: []string{"r", "ro"},
			},
			{
				name:           "empty prefix returns full catalog",
				input:          "4d6",
				cursor:         0,
				expectedSyntax: nil, // checked via count
			},
			{
				name:           "no match returns empty",
				input:          "4d6xyz",
				cursor:         6,
				expectedSyntax: []string{},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				filtered := FilterComponents(cat, tt.input, tt.cursor)
				if tt.expectedSyntax == nil {
					if len(filtered) != len(cat) {
						t.Errorf("FilterComponents with empty token returned %d components, want %d", len(filtered), len(cat))
					}
					return
				}

				if len(filtered) != len(tt.expectedSyntax) {
					t.Fatalf("FilterComponents(%q, %d) returned %d items, want %d", tt.input, tt.cursor, len(filtered), len(tt.expectedSyntax))
				}

				for i, want := range tt.expectedSyntax {
					if filtered[i].Syntax != want {
						t.Errorf("filtered[%d].Syntax = %q, want %q", i, filtered[i].Syntax, want)
					}
				}
			})
		}
	})
}
