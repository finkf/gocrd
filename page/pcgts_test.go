package page

import (
	"fmt"
	"image"
	"testing"
)

const testfile = "testdata/kant_aufklaerung_1784_0020.xml"

func withOpenPcGts(path string, f func(p *PcGts)) {
	p, err := OpenPcGts(path)
	if err != nil {
		panic(err)
	}
	f(p)
}

func want(fmt string, got, want interface{}, t *testing.T) {
	t.Helper()
	if got != want {
		t.Fatalf("expected "+fmt+"; got "+fmt, want, got)
	}
}

func TestUnmarshallPcGts(t *testing.T) {
	withOpenPcGts(testfile, func(p *PcGts) {
		want("%d", len(p.Attributes), 3, t)
		want("%q", p.Metadata["Creator"], "OCR-D", t)
		want("%q", p.Page.Type, "content", t)
		want("%d", len(p.Page.PrintSpace.Coords.Points), 4, t)
		want("%d", p.Page.PrintSpace.Coords.Points[0].X, 468, t)
		want("%d", len(p.Page.ReadingOrder.OrderedGroup), 1, t)
		want("%s", p.Page.ReadingOrder.OrderedGroup[0].ID, "ro_1488818689634", t)
		want("%d", len(p.Page.ReadingOrder.OrderedGroup[0].RegionRefIndexed), 4, t)
		want("%d", p.Page.ReadingOrder.OrderedGroup[0].RegionRefIndexed[1].Index, 1, t)
		want("%d", len(p.Page.TextRegion), 4, t)
		want("%d", p.Page.TextRegion[0].ID, "r_1_1", t)
		want("%d", p.Page.TextRegion[0].ID, "r_1_1", t)
		want("%d", len(p.Page.TextRegion[0].Coords.Points), 4, t)
		want("%d", p.Page.TextRegion[0].Coords.Points[3].Y, 337, t)
		want("%d", len(p.Page.TextRegion[0].TextEquiv.PlainText), 0, t)
		want("%d", len(p.Page.TextRegion[0].TextEquiv.Unicode), 1, t)
		want("%s", p.Page.TextRegion[0].TextEquiv.Unicode[0], "( 484 )", t)
		want("%d", len(p.Page.TextRegion[0].TextLine), 1, t)
		want("%d", len(p.Page.TextRegion[0].TextLine[0].BaseLine.Points), 2, t)
		want("%d", p.Page.TextRegion[0].TextLine[0].BaseLine.Points[1].X, 1025, t)
		want("%s", p.Page.TextRegion[0].TextLine[0].ID, "tl_1", t)
		want("%d", len(p.Page.TextRegion[0].TextLine[0].Word), 3, t)
		want("%f", p.Page.TextRegion[0].TextLine[0].Word[2].TextStyle.FontSize, float32(12.0), t)
		want("%d", len(p.Page.TextRegion[0].TextLine[0].Word[1].TextEquiv.Unicode), 1, t)
		want("%f", p.Page.TextRegion[0].TextLine[0].Word[1].TextEquiv.Unicode[0], "484", t)
	})
}

func TestBoundingBox(t *testing.T) {
	withOpenPcGts(testfile, func(p *PcGts) {
		tests := []struct {
			want image.Rectangle
			test Coords
		}{
			{image.Rect(847, 338, 1025, 338),
				p.Page.TextRegion[0].TextLine[0].BaseLine},
			{image.Rect(847, 295, 862, 335),
				p.Page.TextRegion[0].TextLine[0].Word[0].Coords},
		}
		for _, tc := range tests {
			t.Run(fmt.Sprintf("%s", tc.want), func(t *testing.T) {
				want("%s", tc.test.BoundingBox(), tc.want, t)
			})
		}
	})
}
