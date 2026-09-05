package services

import (
	"testing"
)

func TestConcatenateTitleAndYear(t *testing.T) {
	serv := &CatalogServiceImpl{}

	tests := []struct {
		input         string
		expectedTitle string
		expectedYear  string
	}{
		{
			input:         "Toy Story (1995)",
			expectedTitle: "Toy Story",
			expectedYear:  "1995",
		},
		{
			input:         "American President, The (1995)",
			expectedTitle: "The American President",
			expectedYear:  "1995",
		},
		{
			input:         "Shawshank Redemption, The (1994)",
			expectedTitle: "The Shawshank Redemption",
			expectedYear:  "1994",
		},
		{
			input:         "Usual Suspects, The (1995)",
			expectedTitle: "The Usual Suspects",
			expectedYear:  "1995",
		},
		{
			input:         "Postman, The (Postino, Il) (1994)",
			expectedTitle: "The Postman",
			expectedYear:  "1994",
		},
		{
			input:         "Nightmare on Elm Street, A (1984)",
			expectedTitle: "A Nightmare on Elm Street",
			expectedYear:  "1984",
		},
		{
			input:         "Affair to Remember, An (1957)",
			expectedTitle: "An Affair to Remember",
			expectedYear:  "1957",
		},
		{
			input:         "Seven (a.k.a. Se7en) (1995)",
			expectedTitle: "Seven",
			expectedYear:  "1995",
		},
		{
			input:         "Movie Without Year",
			expectedTitle: "Movie Without Year",
			expectedYear:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			title, year := serv.ConcatenateTitleAndYear(tt.input)
			if title != tt.expectedTitle || year != tt.expectedYear {
				t.Errorf("For %q, got title=%q, year=%q; want title=%q, year=%q",
					tt.input, title, year, tt.expectedTitle, tt.expectedYear)
			}
		})
	}
}
