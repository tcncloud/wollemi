package stringify_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/tcncloud/wollemi/domain/stringify"
)

func TestString(t *testing.T) {
	t.Run("stringifies argument", func(t *testing.T) {
		have := stringify.String(GeorgeRRMartin(), 0)

		require.Equal(t, GeorgeRRMartinString(), have)
	})
}

func TestWrite(t *testing.T) {
	t.Run("writes stringified argument", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)

		err := stringify.Write(buf, GeorgeRRMartin(), 0)
		require.NoError(t, err)
		require.Equal(t, GeorgeRRMartinString(), buf.String())
	})
}

func GeorgeRRMartin() *Author {
	return &Author{
		Name: "George R. R. Martin",
		Born: time.Date(1948, 9, 20, 0, 0, 0, 0, time.UTC),
		Books: []*Book{{
			Title:     "A Game Of Thrones",
			Published: time.Date(1996, 8, 1, 0, 0, 0, 0, time.UTC),
		}, {
			Title:     "A Clash Of Kings",
			Published: time.Date(1998, 11, 16, 0, 0, 0, 0, time.UTC),
		}},
		Awards: map[string][]*Award{
			"Hugo": []*Award{{
				Title:    "A Song for Lya",
				Category: BEST_NOVELLA,
				Year:     1975,
			}, {
				Title:    "Sandkings",
				Category: BEST_NOVELLETE,
				Year:     1980,
			}},
			"Locus": []*Award{{
				Title:    "A Storm of Swords",
				Category: BEST_FANTASY_NOVEL,
				Year:     2003,
			}},
		},
	}
}

func GeorgeRRMartinString() string {
	return `&stringify_test.Author{
	Name: "George R. R. Martin",
	Born: "1948-09-20 00:00:00 +0000 UTC",
	Books: []*stringify_test.Book{
		&stringify_test.Book{
			Title: "A Game Of Thrones",
			Published: "1996-08-01 00:00:00 +0000 UTC",
		},
		&stringify_test.Book{
			Title: "A Clash Of Kings",
			Published: "1998-11-16 00:00:00 +0000 UTC",
		},
	},
	Awards: map[string][]*stringify_test.Award{
		"Hugo":[]*stringify_test.Award{
			&stringify_test.Award{
				Title:    "A Song for Lya",
				Year:     1975,
			},
			&stringify_test.Award{
				Title:    "Sandkings",
				Category: stringify_test.BEST_NOVELLETE,
				Year:     1980,
			},
		},
		"Locus":[]*stringify_test.Award{
			&stringify_test.Award{
				Title:    "A Storm of Swords",
				Category: stringify_test.BEST_FANTASY_NOVEL,
				Year:     2003,
			},
		},
	},
}`
}

type Category int32

func (x Category) String() string {
	return Category_name[int32(x)]
}

const (
	BEST_NOVELLA            Category = 0
	BEST_NOVELLETE          Category = 1
	BEST_SHORT_STORY        Category = 2
	BEST_LONG_FICTION       Category = 3
	BEST_FOREIGN_NOVEL      Category = 4
	BEST_FANTASY_NOVEL      Category = 5
	BEST_ORIGINAL_ANTHOLOGY Category = 6
)

var Category_name = map[int32]string{
	0: "BEST_NOVELLA",
	1: "BEST_NOVELLETE",
	2: "BEST_SHORT_STORY",
	3: "BEST_LONG_FICTION",
	4: "BEST_FOREIGN_NOVEL",
	5: "BEST_FANTASY_NOVEL",
	6: "BEST_ORIGINAL_ANTHOLOGY",
}

type Author struct {
	Name   string
	Born   time.Time
	Books  []*Book
	Awards map[string][]*Award
}

type Book struct {
	Title     string
	Published time.Time
}

type Award struct {
	Title    string
	Category Category
	Year     uint32
}
