package mets // import "github.com/finkf/gocrd/mets"

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

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
// It returns a list of matching files.
func (m Mets) FindFileGrp(use string) []File {
	return m.Find(Match{Use: use})
}

// Find returns a list of matching files. Empty fields in the
// given match are ignored for the matching.
func (m Mets) Find(match Match) []File {
	var fs []File
	for i := match.xpath().Iter(m.root); i.Next(); {
		fs = append(fs, newFileFromNode(i.Node()))
	}
	return fs
}

// Match is used to match files.
// If a field is the empty string it is ignored for the matching.
type Match struct {
	Use, FileID, MIMEType, GroupID string
}

func (m Match) String() string {
	return fmt.Sprintf("{%q,%q,%q}", m.Use, m.FileID, m.MIMEType)
}

func (m Match) xpath() *xmlpath.Path {
	xpath := "/mets/fileSec/fileGrp/file"
	if m.Use != "" {
		xpath = fmt.Sprintf("/mets/fileSec/fileGrp[@USE=%q]/file", m.Use)
	}
	if m.FileID != "" {
		xpath += fmt.Sprintf("[@ID=%q]", m.FileID)
	}
	if m.MIMEType != "" {
		xpath += fmt.Sprintf("[@MIMETYPE=%q]", m.MIMEType)
	}
	if m.GroupID != "" {
		xpath += fmt.Sprintf("[@GROUPID=%q]", m.GroupID)
	}
	return xmlpath.MustCompile(xpath)
}

// FLocat represents a mets:FLocat of a mets:file entry.
type FLocat struct {
	Type, URL string
}

// Open opens a reader to the FLocat.
// Currently only URL's are supported.
func (f FLocat) Open() (io.ReadCloser, error) {
	if f.Type != "URL" {
		return nil, fmt.Errorf("open: unsupported LOCTYPE: %s", f.Type)
	}
	if strings.HasPrefix(f.URL, "file://") {
		return os.Open(f.URL[7:])
	}
	r, err := http.Get(f.URL)
	if err != nil {
		return nil, fmt.Errorf("open: %v", err)
	}
	return r.Body, nil
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
