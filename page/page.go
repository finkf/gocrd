package page // import "github.com/finkf/gocrd/page"

import (
	"fmt"
	"os"
	"sort"
	"strconv"

	"launchpad.net/xmlpath"
)

const (
	// MIMEType defines the mime-type of page XML files.
	// See: https://github.com/PRImA-Research-Lab/PAGE-XML
	MIMEType = "application/alto+xml"
)

// XPath helpers
var (
	indexXPath            = xmlpath.MustCompile("@index")
	regionRefXPath        = xmlpath.MustCompile("@regionRef")
	idXPath               = xmlpath.MustCompile("@id")
	regionRefIndexedXPath = xmlpath.MustCompile("/PcGts/Page/ReadingOrder/*/RegionRefIndexed")
	wordsXPath            = xmlpath.MustCompile("./Word")
	// check interface types
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
	Polygon() (Polygon, error)
}

// Page represents an open page XML file.
type Page struct {
	path string
	root *xmlpath.Node
}

// Open opens a page XML file
func Open(path string) (Page, error) {
	in, err := os.Open(path)
	if err != nil {
		return Page{}, err
	}
	defer func() { _ = in.Close() }()
	root, err := xmlpath.Parse(in)
	if err != nil {
		return Page{}, err
	}
	return Page{path, root}, nil
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
	if m.RegionID != "" {
		suffix = fmt.Sprintf("/TextRegion[@id=%q]%s", m.RegionID, suffix)
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

// Find searches for a given {region,line,word}-ID in the PAGE-XML
// (IDs are assumed to be unique).
func (p Page) Find(m Match) (TextRegion, bool) {
	return m.find(p.root)
}

// Regions returns a slice with all RegionRefIndexed elements
func (p Page) Regions() []Region {
	var regions []Region
	for i := regionRefIndexedXPath.Iter(p.root); i.Next(); {
		region, err := newRegion(p.root, i.Node())
		if err != nil { // skip erroneous nodes
			continue
		}
		regions = append(regions, region)
	}
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

// Region defines a text region in the page XML file.
type Region struct {
	id    string
	root  *xmlpath.Node
	index int
}

// Lines Returns all lines in a region.
func (r Region) Lines() []Line {
	var lines []Line
	for i := linesXPath(r.id).Iter(r.root); i.Next(); {
		node := i.Node()
		lines = append(lines, Line{node, idFromNode(node)})
	}
	return lines
}

// ID returns the region's ID.
func (w Region) ID() string {
	return w.id
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

func newRegion(root, node *xmlpath.Node) (Region, error) {
	region := Region{root: root}
	str, ok := indexXPath.String(node)
	if ok {
		index, err := strconv.Atoi(str)
		if err != nil {
			return Region{}, err
		}
		region.index = index
	}
	str, ok = regionRefXPath.String(node)
	if ok {
		region.id = str
	}
	return region, nil
}

// Line represents a line of text in the page XML file.
type Line struct {
	node *xmlpath.Node
	id   string
}

// ID returns the line's ID.
func (w Line) ID() string {
	return w.id
}

// TextEquivUnicodeAt returns the i-th TextEquiv/Unicode entry
// (indexing is zero-based).
func (l Line) TextEquivUnicodeAt(pos int) (string, bool) {
	return textEquivUnicodeXPath(pos).String(l.node)
}

// Words returns all words in a line.
func (l Line) Words() []Word {
	var words []Word
	for i := wordsXPath.Iter(l.node); i.Next(); {
		node := i.Node()
		words = append(words, Word{node, idFromNode(node)})
	}
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

// Word represents a word on a line.
type Word struct {
	node *xmlpath.Node
	id   string
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
