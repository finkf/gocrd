package mets // import "github.com/finkf/gocrd/mets"

import (
	"fmt"
	"os"

	"launchpad.net/xmlpath"
)

var (
	mimeTypeXPath = xmlpath.MustCompile("@MIMETYPE")
	idXPath       = xmlpath.MustCompile("@ID")
	hrefXPath     = xmlpath.MustCompile("@href")
	locTypeXPath  = xmlpath.MustCompile("@LOCTYPE")
	flocatXPath   = xmlpath.MustCompile("./FLocat")
)

func fileGrpUseXPath(use string) *xmlpath.Path {
	return xmlpath.MustCompile(fmt.Sprintf("/mets/fileSec/fileGrp[@USE=%q]/file", use))
}

// Mets represents an open METS file.
type Mets struct {
	path string
	root *xmlpath.Node
}

// Open opens a Mets structure from a given path.
func Open(path string) (Mets, error) {
	in, err := os.Open(path)
	if err != nil {
		return Mets{}, err
	}
	defer func() { _ = in.Close() }()
	root, err := xmlpath.Parse(in)
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
func (m Mets) FindFileGrp(use string) []File {
	var fs []File
	for i := fileGrpUseXPath(use).Iter(m.root); i.Next(); {
		fs = append(fs, newFileFromNode(i.Node()))
	}
	return fs
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

func newFileFromNode(n *xmlpath.Node) File {
	var file File
	str, ok := mimeTypeXPath.String(n)
	if ok {
		file.MIMEType = str
	}
	str, ok = idXPath.String(n)
	if ok {
		file.ID = str
	}
	file.FLocat = newFLocatFromNode(n)
	return file
}

func newFLocatFromNode(n *xmlpath.Node) FLocat {
	i := flocatXPath.Iter(n)
	if !i.Next() {
		return FLocat{}
	}
	n = i.Node()
	var flocat FLocat
	str, ok := hrefXPath.String(n)
	if ok {
		flocat.URL = str
	}
	str, ok = locTypeXPath.String(n)
	if ok {
		flocat.Type = str
	}
	return flocat
}

// <mets:file ID="PAGE_0020_ALTO" MIMETYPE="application/alto+xml">
//     <mets:FLocat LOCTYPE="URL" xlink:href="https://github.com/OCR-D/assets/raw/master/data/
