package mets // import "github.com/finkf/gocrd/mets"

import (
	"fmt"
	"os"

	"github.com/antchfx/xmlquery"
)

// Mets represents an open METS file.
type Mets struct {
	path string
	root *xmlquery.Node
}

// Open opens a Mets structure from a given path.
func Open(path string) (Mets, error) {
	in, err := os.Open(path)
	if err != nil {
		return Mets{}, err
	}
	defer func() { _ = in.Close() }()
	root, err := xmlquery.Parse(in)
	if err != nil {
		return Mets{}, err
	}
	return Mets{
		path: path,
		root: root,
	}, nil
}

// FindFileGrp searches for a file group with the given USE flag.
// It returns a list of files and true if a file group
// with the given USE flag was found.
// Note that the list of files can be empty even if true is returned.
func (m Mets) FindFileGrp(use string) ([]File, bool) {
	grp := xmlquery.FindOne(m.root, xpathFileGrpUse(use))
	if grp == nil {
		return nil, false
	}
	var fs []File
	files := xmlquery.Find(grp, "./mets:file")
	for _, n := range files {
		fs = append(fs, newFileFromNode(n))
	}
	return fs, true
}

func xpathFileGrpUse(use string) string {
	return fmt.Sprintf("/mets:mets/mets:fileSec/mets:fileGrp[@USE=%q]", use)
}

// FLocat represents a mets:FLocat of a mets:file entry.
type FLocat struct {
	Type, URL string
}

// File represents a mets:file entry
type File struct {
	MIMEType, ID string
	FLocat       FLocat
}

func newFileFromNode(n *xmlquery.Node) File {
	var file File
	if n == nil {
		return file
	}
	i := xmlquery.CreateXPathNavigator(n)
	for i.MoveToNextAttribute() {
		switch i.LocalName() {
		case "ID":
			file.ID = i.Value()
		case "MIMETYPE":
			file.MIMEType = i.Value()
		}
	}
	file.FLocat = newFLocatFromNode(xmlquery.FindOne(n, "./mets:FLocat"))
	return file
}

func newFLocatFromNode(n *xmlquery.Node) FLocat {
	var flocat FLocat
	if n == nil {
		return flocat
	}
	i := xmlquery.CreateXPathNavigator(n)
	for i.MoveToNextAttribute() {
		switch i.LocalName() {
		case "LOCTYPE":
			flocat.Type = i.Value()
		case "href":
			flocat.URL = i.Value()
		}
	}
	return flocat
}

// <mets:file ID="PAGE_0020_ALTO" MIMETYPE="application/alto+xml">
//     <mets:FLocat LOCTYPE="URL" xlink:href="https://github.com/OCR-D/assets/raw/master/data/
