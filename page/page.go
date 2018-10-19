package page // import "github.com/finkf/gocrd/page"

import (
	"fmt"
	"image"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"launchpad.net/xmlpath"
)

const (
	// MIMEType defines the mime-type of page XML files.
	// See: https://github.com/PRImA-Research-Lab/PAGE-XML
	MIMEType = "application/alto+xml"
)

// XPath helpers
var (
	coordsXPath           = xmlpath.MustCompile("./Coords/@points")
	indexXPath            = xmlpath.MustCompile("@index")
	regionRefXPath        = xmlpath.MustCompile("@regionRef")
	idXPath               = xmlpath.MustCompile("@id")
	regionRefIndexedXPath = xmlpath.MustCompile("/PcGts/Page/ReadingOrder/*/RegionRefIndexed")
	wordsXPath            = xmlpath.MustCompile("./Word")
	// check interface types
	_ TextRegion = Page{}
	_ TextRegion = Region{}
	_ TextRegion = Line{}
	_ TextRegion = Word{}
)

func textEquivUnicodeXPath(i int) *xmlpath.Path {
	return xmlpath.MustCompile(fmt.Sprintf("./TextEquiv[%d]/Unicode", i+1))
}

func linesXPath(id string) *xmlpath.Path {
	return xmlpath.MustCompile(fmt.Sprintf("/PcGts/Page/TextRegion[@id=%q]/TextLine", id))
}

func regionXPath(id string) *xmlpath.Path {
	return xmlpath.MustCompile(fmt.Sprintf("/PcGts/Page/TextRegion[@id=%q]", id))
}

func idFromNode(node *xmlpath.Node) string {
	id, ok := idXPath.String(node)
	if !ok {
		return ""
	}
	return id
}

// TextRegion defines an interface for abstract
// text regions in a PAGE-XML document.
type TextRegion interface {
	ID() string
	TextEquivUnicodeAt(int) (string, bool)
	EachSubRegion(func(TextRegion))
	Polygon() (Polygon, error)
}

// Match is used to match text regions.
// If any of the IDs is the empty string,
// the according region is ignored.
type Match struct {
	RegionID, LineID, WordID string
}

func (m Match) xpath() *xmlpath.Path {
	suffix := ""
	if m.WordID != "" {
		suffix = fmt.Sprintf("/Word[@id=%q]", m.WordID)
	}
	if m.LineID != "" {
		suffix = fmt.Sprintf("/TextLine[@id=%q]%s", m.LineID, suffix)
	} else if suffix != "" {
		suffix = "/*" + suffix
	}
	if m.RegionID != "" && suffix != "" {
		suffix = fmt.Sprintf("/TextRegion[@id=%q]%s", m.RegionID, suffix)
	} else if m.RegionID != "" {
		suffix = fmt.Sprintf("/ReadingOrder/*/RegionRefIndexed[@regionRef=%q]", m.RegionID)
	} else if suffix != "" {
		suffix = "/*" + suffix
	}
	return xmlpath.MustCompile("/PcGts/Page" + suffix)
}

func (m Match) find(root *xmlpath.Node) (TextRegion, bool) {
	if i := m.xpath().Iter(root); i.Next() {
		if m.WordID != "" {
			return Word{i.Node(), idFromNode(i.Node())}, true
		}
		if m.LineID != "" {
			return Line{i.Node(), idFromNode(i.Node())}, true
		}
		if m.RegionID != "" {
			r, err := newRegion(root, i.Node())
			if err != nil {
				return nil, false
			}
			return r, true
		}
	}
	return nil, false
}

func (m Match) String() string {
	return fmt.Sprintf("{%q,%q,%q}", m.RegionID, m.LineID, m.WordID)
}

// Page represents an open page XML file.
type Page struct {
	path string
	root *xmlpath.Node
}

// Open opens a page XML file
func Open(path string) (Page, error) {
	is, err := os.Open(path)
	if err != nil {
		return Page{}, err
	}
	defer is.Close()
	root, err := xmlpath.Parse(is)
	if err != nil {
		return Page{}, err
	}
	return Page{path, root}, nil
}

// ID returns the ID of a page.
// The ID of a page is the page's basename.
func (p Page) ID() string {
	return filepath.Base(p.path)
}

// Polygon returns the page's print-space.
func (p Page) Polygon() (Polygon, error) {
	xpath := xmlpath.MustCompile("/PcGts/Page/PrintSpace")
	i := xpath.Iter(p.root)
	if !i.Next() {
		return Polygon{}, fmt.Errorf("missing PrintSpace for %s", p.path)
	}
	return newPolygon(i.Node())
}

// TextEquivUnicodeAt returns the i-th TextEquiv/Unicode entry
// for each of the page's regions, separated by \n\n.
func (p Page) TextEquivUnicodeAt(pos int) (string, bool) {
	var b strings.Builder
	var pre string
	for _, r := range p.Regions() {
		region, ok := r.TextEquivUnicodeAt(pos)
		if !ok {
			return "", false
		}
		b.WriteString(pre)
		b.WriteString(region)
		pre = "\n\n"
	}
	return b.String(), true
}

// Find searches for a given {region,line,word}-ID in the PAGE-XML
// (IDs are assumed to be unique).
func (p Page) Find(m Match) (TextRegion, bool) {
	return m.find(p.root)
}

// Regions returns a slice with all RegionRefIndexed elements
func (p Page) Regions() []Region {
	var regions []Region
	iterate(regionRefIndexedXPath, p.root, func(node *xmlpath.Node) {
		region, err := newRegion(p.root, node)
		if err != nil { // skip erroneous nodes
			return
		}
		regions = append(regions, region)
	})
	sort.Slice(regions, func(i, j int) bool {
		return regions[i].index < regions[j].index
	})
	return regions
}

// FindRegionByID returns the region with the given ID.
func (p Page) FindRegionByID(id string) (Region, bool) {
	for _, region := range p.Regions() {
		if region.id == id {
			return region, true
		}
	}
	return Region{}, false
}

// EachSubRegion calls the given callback function for
// each sub region (region) of the page.
func (p Page) EachSubRegion(f func(TextRegion)) {
	for _, r := range p.Regions() {
		f(r)
	}
}

// Region defines a text region in the page XML file.
type Region struct {
	id    string
	root  *xmlpath.Node
	index int
}

// Lines Returns all lines in a region.
func (r Region) Lines() []Line {
	var lines []Line
	iterate(linesXPath(r.id), r.root, func(node *xmlpath.Node) {
		lines = append(lines, Line{node, idFromNode(node)})
	})
	return lines
}

// EachSubRegion calls the given callback function for
// each sub region (line) of the region.
func (r Region) EachSubRegion(f func(TextRegion)) {
	iterate(linesXPath(r.id), r.root, func(node *xmlpath.Node) {
		f(Line{node, idFromNode(node)})
	})
}

// ID returns the region's ID.
func (r Region) ID() string {
	return r.id
}

// FindLineByID searches for a line with the given ID.
func (r Region) FindLineByID(id string) (Line, bool) {
	for _, line := range r.Lines() {
		if line.id == id {
			return line, true
		}
	}
	return Line{}, false
}

// TextEquivUnicodeAt returns the i-th TextEquiv/Unicode entry
// (indexing is zero-based).
func (r Region) TextEquivUnicodeAt(pos int) (string, bool) {
	if i := regionXPath(r.id).Iter(r.root); i.Next() {
		return textEquivUnicodeXPath(pos).String(i.Node())
	}
	return "", false
}

// Polygon returns the region's polygon of coordinates.
func (r Region) Polygon() (Polygon, error) {
	if i := regionXPath(r.id).Iter(r.root); i.Next() {
		return newPolygon(i.Node())
	}
	return nil, fmt.Errorf("invalid region: %s", r.id)
}

func newRegion(root, node *xmlpath.Node) (Region, error) {
	region := Region{root: root}
	str, ok := indexXPath.String(node)
	if !ok {
		return Region{}, fmt.Errorf("invalid region: missing index")
	}
	index, err := strconv.Atoi(str)
	if err != nil {
		return Region{}, fmt.Errorf("invalid region: %v", err)
	}
	region.index = index
	str, ok = regionRefXPath.String(node)
	if !ok {
		return Region{}, fmt.Errorf("invalid region: missing id")
	}
	region.id = str
	return region, nil
}

// Line represents a line of text in the page XML file.
type Line struct {
	node *xmlpath.Node
	id   string
}

// EachSubRegion calls the given callback function for
// each sub region (word) of the region.
func (l Line) EachSubRegion(f func(TextRegion)) {
	iterate(wordsXPath, l.node, func(node *xmlpath.Node) {
		f(Word{node, idFromNode(node)})
	})
}

// ID returns the line's ID.
func (l Line) ID() string {
	return l.id
}

// TextEquivUnicodeAt returns the i-th TextEquiv/Unicode entry
// (indexing is zero-based).
func (l Line) TextEquivUnicodeAt(pos int) (string, bool) {
	return textEquivUnicodeXPath(pos).String(l.node)
}

// Words returns all words in a line.
func (l Line) Words() []Word {
	var words []Word
	iterate(wordsXPath, l.node, func(node *xmlpath.Node) {
		words = append(words, Word{node, idFromNode(node)})
	})
	return words
}

// FindWordByID searches for a line with the given ID.
func (l Line) FindWordByID(id string) (Word, bool) {
	for _, word := range l.Words() {
		if word.id == id {
			return word, true
		}
	}
	return Word{}, false
}

// Polygon returns the line's polygon of coordinates.
func (l Line) Polygon() (Polygon, error) {
	return newPolygon(l.node)
}

// Word represents a word on a line.
type Word struct {
	node *xmlpath.Node
	id   string
}

// EachSubRegion calls the given callback function for
// each sub region (Glyph) of the region.
func (w Word) EachSubRegion(f func(TextRegion)) {
}

// ID returns the word's ID.
func (w Word) ID() string {
	return w.id
}

// TextEquivUnicodeAt returns the i-th TextEquiv/Unicode element
// (the indexing is zero-based).
func (w Word) TextEquivUnicodeAt(pos int) (string, bool) {
	return textEquivUnicodeXPath(pos).String(w.node)
}

// Polygon returns the word's polygon of coordinates.
func (w Word) Polygon() (Polygon, error) {
	return newPolygon(w.node)
}

// Polygon is used to represent the polygons of
// <Coords points='...'/> points in the PAGE-XML.
type Polygon []image.Point

// Rectangle returns the bounding rectangle of the polygon.
func (p Polygon) Rectangle() image.Rectangle {
	minx := math.MaxInt64
	maxx := math.MinInt64
	miny := math.MaxInt64
	maxy := math.MinInt64
	for _, p := range p {
		if p.X < minx {
			minx = p.X
		}
		if p.Y < miny {
			miny = p.Y
		}
		if p.X > maxx {
			maxx = p.X
		}
		if p.Y > maxy {
			maxy = p.Y
		}
	}
	return image.Rect(int(minx), int(miny), int(maxx), int(maxy))
}

// <Coords points="846,294 1026,294 1026,337 846,337"/>
func newPolygon(node *xmlpath.Node) (Polygon, error) {
	psstr, ok := coordsXPath.String(node)
	if !ok {
		return nil, fmt.Errorf("invalid coordinates: missing")
	}
	var points []image.Point
	ps := strings.Split(psstr, " ")
	if len(ps) < 2 {
		return nil, fmt.Errorf("invalid coordinates: %q", psstr)
	}
	for _, p := range ps {
		point := strings.Split(p, ",")
		if len(point) != 2 {
			return nil, fmt.Errorf("invalid coordinates: invalid point: %q", p)
		}
		x, err := strconv.Atoi(point[0])
		if err != nil {
			return nil, err
		}
		y, err := strconv.Atoi(point[1])
		if err != nil {
			return nil, err
		}
		points = append(points, image.Point{X: x, Y: y})
	}
	return points, nil
}

func iterate(xpath *xmlpath.Path, root *xmlpath.Node, f func(*xmlpath.Node)) {
	for i := xpath.Iter(root); i.Next(); {
		f(i.Node())
	}
}
