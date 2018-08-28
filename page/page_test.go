package page

import "testing"

func TestFindRegionByRefID(t *testing.T) {
	tests := []struct {
		id   string
		find bool
	}{
		{"invalid-ref-id", false},
		{"r_1_1", true},
		{"r_2_1", true},
		{"r_2_2", true},
		{"r_2_3", true},
		{"r_1_2", false},
	}
	for _, tc := range tests {
		t.Run(tc.id, func(t *testing.T) {
			page, err := Open("testdata/kant_aufklaerung_1784_0020.xml")
			if err != nil {
				t.Fatalf("got error: %v", err)
			}
			region, ok := page.FindRegionByID(tc.id)
			if ok != tc.find {
				t.Fatalf("expected ok=%t; got ok=%t", tc.find, ok)
			}
			if tc.find && region.ID != tc.id {
				t.Fatalf("expected %s; got %s", region.ID, tc.id)
			}
		})
	}
}

func TestFindLineByID(t *testing.T) {
	tests := []struct {
		refID, lineID string
		find          bool
	}{
		{"r_1_1", "invalid-line-id", false},
		{"r_1_1", "tl_1", true},
		{"r_1_1", "tl_2", false},
		{"r_2_1", "tl_1", false},
		{"r_2_1", "tl_2", true},
	}
	for _, tc := range tests {
		t.Run(tc.refID+" "+tc.lineID, func(t *testing.T) {
			page, err := Open("testdata/kant_aufklaerung_1784_0020.xml")
			if err != nil {
				t.Fatalf("got error: %v", err)
			}
			r, _ := page.FindRegionByID(tc.refID)
			l, ok := r.FindLineByID(tc.lineID)
			if tc.find != ok {
				t.Fatalf("expected ok=%t; got ok=%t", tc.find, ok)
			}
			if tc.find && l.ID != tc.lineID {
				t.Fatalf("expected %s; got %s", tc.lineID, l.ID)
			}
		})
	}
}

// func TestLineTextEquivUnicode(t *testing.T) {
// 	tests := []struct {
// 		refID, lineID, unicode string
// 	}{
// 		{"r_1_1", "tl_1", "( 484 )"},
// 		{"r_2_1", "tl_13", "dienen."},
// 	}
// 	for _, tc := range tests {
// 		t.Run(tc.refID+" "+tc.lineID, func(t *testing.T) {
// 			page, err := Open("testdata/kant_aufklaerung_1784_0020.xml")
// 			if err != nil {
// 				t.Fatalf("got error: %v", err)
// 			}
// 			r, _ := page.FindRegionByRefID(tc.refID)
// 			l, _ := r.FindLineByID(tc.lineID)
// 			if got, _ := l.TextEquivUnicodeAt(0); got != tc.unicode {
// 				t.Fatalf("expceted %q; got %q", tc.unicode, got)
// 			}
// 		})
// 	}
// }

// func TestFindWordByID(t *testing.T) {
// 	tests := []struct {
// 		refID, lineID, wordID, word string
// 		find                        bool
// 	}{
// 		{"r_1_1", "tl_1", "invalid-word-id", "", false},
// 		{"r_1_1", "tl_1", "w_w1aab1b1b2b1b1ab1", "(", true},
// 		{"r_2_1", "tl_2", "w_w1aab1b3b2b1b1ab1", "gewiegelt", true},
// 	}
// 	for _, tc := range tests {
// 		t.Run(tc.refID+" "+tc.lineID, func(t *testing.T) {
// 			page, err := Open("testdata/kant_aufklaerung_1784_0020.xml")
// 			if err != nil {
// 				t.Fatalf("got error: %v", err)
// 			}
// 			r, _ := page.FindRegionByRefID(tc.refID)
// 			l, _ := r.FindLineByID(tc.lineID)
// 			w, ok := l.FindWordByID(tc.wordID)
// 			if tc.find != ok {
// 				t.Fatalf("expected ok=%t; got ok=%t", tc.find, ok)
// 			}
// 			if tc.find && w.ID != tc.wordID {
// 				t.Fatalf("expected %s; got %s", tc.wordID, w.ID)
// 			}
// 			if tc.find {
// 				if got, _ := w.TextEquivUnicodeAt(0); got != tc.word {
// 					t.Fatalf("expected %s; got %s", tc.word, got)
// 				}
// 			}
// 		})
// 	}
// }
