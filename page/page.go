package page // import "github.com/finkf/gocrd/page"

import (
	"fmt"
	"os"
	"sort"
	"strconv"

	xmlquery "github.com/antchfx/xmlquery"
	"github.com/pkg/errors"
)

const (
	// MIMEType defines the mime-type of page XML files.
	// See: https://github.com/PRImA-Research-Lab/PAGE-XML
	MIMEType = "application/alto+xml"
)

// Page represents an open page XML file.
type Page struct {
	path    string
	root    *xmlquery.Node
	regions []Region
}

// Open opens a page XML file
func Open(path string) (Page, error) {
	in, err := os.Open(path)
	if err != nil {
		return Page{}, err
	}
	defer func() { _ = in.Close() }()
	root, err := xmlquery.Parse(in)
	if err != nil {
		return Page{}, err
	}
	p := Page{path, root, nil}
	return p, errors.Wrapf(p.readRegions(), "invalid page XML: %s", path)
}

func (p *Page) readRegions() error {
	regions := xmlquery.Find(p.root, "/PcGts/Page/ReadingOrder/*/RegionRefIndexed")
	for _, r := range regions {
		region, err := newRegion(p.root, r)
		if err != nil {
			return err
		}
		p.regions = append(p.regions, region)
	}
	// sort by region.index
	sort.Slice(p.regions, func(i, j int) bool {
		return p.regions[i].index < p.regions[j].index
	})
	return nil
}

// FindRegionsByGroupID returns all regions with the given group ID.
func (p Page) FindRegionsByGroupID(groupID string) []Region {
	var regions []Region
	for _, region := range p.regions {
		if region.GroupID == groupID {
			regions = append(regions, region)
		}
	}
	return regions
}

// FindRegionByRefID returns the region with the given refID.
func (p Page) FindRegionByRefID(refID string) (Region, bool) {
	for _, region := range p.regions {
		if region.RefID == refID {
			return region, true
		}
	}
	return Region{}, false
}

// Regions returns all regions in the page XML file.
func (p Page) Regions() []Region {
	return p.regions
}

// Region defines a text region in the page XML file.
type Region struct {
	GroupID, RefID string
	node           *xmlquery.Node
	index          int
}

// Lines Returns all lines in a region.
func (r Region) Lines() []TextLine {
	tls := xmlquery.Find(r.node, "./TextLine")
	var lines []TextLine
	for _, tl := range tls {
		lines = append(lines, TextLine{tl, getID(tl)})
	}
	return lines
}

// FindLineByID searches for a line with the given ID.
func (r Region) FindLineByID(id string) (TextLine, bool) {
	for _, line := range r.Lines() {
		if line.ID == id {
			return line, true
		}
	}
	return TextLine{}, false
}

// TextEquivUnicodeAt returns the i-th TextEquiv/Unicode entry
// (indexing is zero-based).
func (r Region) TextEquivUnicodeAt(i int) (string, bool) {
	return textEquivTypeUnicodeAt(r.node, i)
}

// newRegion creates a new region with the according
// GroupID, RefID and index. The function searches for the
// according TextRegion node in the page XML file.
func newRegion(root, node *xmlquery.Node) (Region, error) {
	region := Region{}
	i := xmlquery.CreateXPathNavigator(node)
	for i.MoveToNextAttribute() {
		switch i.LocalName() {
		case "index":
			index, err := strconv.Atoi(i.Value())
			if err != nil {
				return Region{}, errors.Wrapf(err, "invalid index: %s", i.Value())
			}
			region.index = index
		case "regionRef":
			region.RefID = i.Value()
		}
	}
	textRegionNode := xmlquery.FindOne(
		root, fmt.Sprintf("/PcGts/Page/TextRegion[@id=%q]", region.RefID))
	if textRegionNode == nil {
		return Region{}, fmt.Errorf("cannot find Region id: %s", region.RefID)
	}
	region.node = textRegionNode
	i = xmlquery.CreateXPathNavigator(node)
	if i.MoveToParent() && i.LocalName() == "OrderedGroup" {
		for i.MoveToNextAttribute() {
			if i.LocalName() == "id" {
				region.GroupID = i.Value()
				break
			}
		}
	}
	return region, nil
}

// TextLine represents a line of text in the page XML file.
type TextLine struct {
	node *xmlquery.Node
	ID   string
}

// TextEquivUnicodeAt returns the i-th TextEquiv/Unicode element
// (the indexing is zero-based).
func (l TextLine) TextEquivUnicodeAt(i int) (string, bool) {
	return textEquivTypeUnicodeAt(l.node, i)
}

// Words returns all words in a line.
func (l TextLine) Words() []Word {
	wds := xmlquery.Find(l.node, "./Word")
	var words []Word
	for _, w := range wds {
		words = append(words, Word{w, getID(w)})
	}
	return words
}

// FindWordByID searches for a line with the given ID.
func (l TextLine) FindWordByID(id string) (Word, bool) {
	for _, word := range l.Words() {
		if word.ID == id {
			return word, true
		}
	}
	return Word{}, false
}

// Word represents a word on a line.
type Word struct {
	node *xmlquery.Node
	ID   string
}

// TextEquivUnicodeAt returns the i-th TextEquiv/Unicode element
// (the indexing is zero-based).
func (w Word) TextEquivUnicodeAt(i int) (string, bool) {
	return textEquivTypeUnicodeAt(w.node, i)
}

func getID(node *xmlquery.Node) string {
	i := xmlquery.CreateXPathNavigator(node)
	for i.MoveToNextAttribute() {
		switch i.LocalName() {
		case "id":
			return i.Value()
		}
	}
	return ""
}

func textEquivTypeUnicodeAt(equiv *xmlquery.Node, i int) (string, bool) {
	u := xmlquery.FindOne(equiv, fmt.Sprintf("./TextEquiv[%d]/Unicode", i+1))
	if u == nil {
		return "", false
	}
	if u.FirstChild == nil || u.FirstChild.Type != xmlquery.TextNode {
		return "", false
	}
	return u.FirstChild.Data, true
}
