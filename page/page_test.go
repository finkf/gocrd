package page

import "testing"

func TestFindRegionByRefID(t *testing.T) {
	tests := []struct {
		refID string
		find  bool
	}{
		{"invalid-ref-id", false},
		{"r_1_1", true},
		{"r_2_1", true},
		{"r_2_2", true},
		{"r_2_3", true},
		{"r_1_2", false},
	}
	for _, tc := range tests {
		t.Run(tc.refID, func(t *testing.T) {
			page, err := Open("testdata/kant_aufklaerung_1784_0020.xml")
			if err != nil {
				t.Fatalf("got error: %v", err)
			}
			region, ok := page.FindRegionByRefID(tc.refID)
			if ok != tc.find {
				t.Fatalf("expected ok=%t; got ok=%t", tc.find, ok)
			}
			if tc.find && region.RefID != tc.refID {
				t.Fatalf("expected %s; got %s", region.RefID, tc.refID)
			}
		})
	}
}

func TestFindRegionsByGroupID(t *testing.T) {
	tests := []struct {
		groupID string
		len     int
	}{
		{"invalid-ref-id", 0},
		{"ro_1488818689634", 4},
	}
	for _, tc := range tests {
		t.Run(tc.groupID, func(t *testing.T) {
			page, err := Open("testdata/kant_aufklaerung_1784_0020.xml")
			if err != nil {
				t.Fatalf("got error: %v", err)
			}
			regions := page.FindRegionsByGroupID(tc.groupID)
			if got := len(regions); got != tc.len {
				t.Fatalf("expected %d; got %d", tc.len, got)
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
			r, _ := page.FindRegionByRefID(tc.refID)
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

func TestLineTextEquivUnicode(t *testing.T) {
	tests := []struct {
		refID, lineID, unicode string
	}{
		{"r_1_1", "tl_1", "( 484 )"},
		{"r_2_1", "tl_13", "dienen."},
	}
	for _, tc := range tests {
		t.Run(tc.refID+" "+tc.lineID, func(t *testing.T) {
			page, err := Open("testdata/kant_aufklaerung_1784_0020.xml")
			if err != nil {
				t.Fatalf("got error: %v", err)
			}
			r, _ := page.FindRegionByRefID(tc.refID)
			l, _ := r.FindLineByID(tc.lineID)
			if got, _ := l.TextEquivUnicodeAt(0); got != tc.unicode {
				t.Fatalf("expceted %q; got %q", tc.unicode, got)
			}
		})
	}
}
