package android

import (
	"encoding/xml"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Element is a flattened, model-friendly view of a uiautomator node. Bounds and
// Center are in true device pixels, so Center can be passed straight to Tap.
type Element struct {
	Text       string `json:"text,omitempty"`
	Desc       string `json:"content_desc,omitempty"`
	ResourceID string `json:"resource_id,omitempty"`
	Class      string `json:"class,omitempty"`
	Clickable  bool   `json:"clickable"`
	Focused    bool   `json:"focused,omitempty"`
	Bounds     Bounds `json:"bounds"`
	Center     Point  `json:"center"`
}

// Bounds is the pixel rectangle of an element: [X1,Y1] top-left, [X2,Y2] bottom-right.
type Bounds struct {
	X1 int `json:"x1"`
	Y1 int `json:"y1"`
	X2 int `json:"x2"`
	Y2 int `json:"y2"`
}

// Point is a tap coordinate in true device pixels.
type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// xmlNode mirrors the <node> element emitted by `uiautomator dump`.
type xmlNode struct {
	Text      string    `xml:"text,attr"`
	Desc      string    `xml:"content-desc,attr"`
	Resource  string    `xml:"resource-id,attr"`
	Class     string    `xml:"class,attr"`
	Clickable string    `xml:"clickable,attr"`
	Focused   string    `xml:"focused,attr"`
	Bounds    string    `xml:"bounds,attr"`
	Children  []xmlNode `xml:"node"`
}

type xmlHierarchy struct {
	XMLName xml.Name  `xml:"hierarchy"`
	Nodes   []xmlNode `xml:"node"`
}

var boundsRe = regexp.MustCompile(`\[(-?\d+),(-?\d+)\]\[(-?\d+),(-?\d+)\]`)

// ParseHierarchy parses uiautomator XML into a flat list of interesting
// elements: those carrying text, a content-description, a resource id, or that
// are clickable. Pure-layout container nodes are dropped to keep the output
// small and useful for choosing tap targets.
func ParseHierarchy(xmlData string) ([]Element, error) {
	var h xmlHierarchy
	if err := xml.Unmarshal([]byte(xmlData), &h); err != nil {
		return nil, fmt.Errorf("parse uiautomator xml: %w", err)
	}
	var out []Element
	for i := range h.Nodes {
		walk(&h.Nodes[i], &out)
	}
	return out, nil
}

func walk(n *xmlNode, out *[]Element) {
	clickable := n.Clickable == "true"
	interesting := strings.TrimSpace(n.Text) != "" ||
		strings.TrimSpace(n.Desc) != "" ||
		strings.TrimSpace(n.Resource) != "" ||
		clickable
	if interesting {
		if b, ok := parseBounds(n.Bounds); ok {
			*out = append(*out, Element{
				Text:       n.Text,
				Desc:       n.Desc,
				ResourceID: n.Resource,
				Class:      n.Class,
				Clickable:  clickable,
				Focused:    n.Focused == "true",
				Bounds:     b,
				Center:     Point{X: (b.X1 + b.X2) / 2, Y: (b.Y1 + b.Y2) / 2},
			})
		}
	}
	for i := range n.Children {
		walk(&n.Children[i], out)
	}
}

func parseBounds(s string) (Bounds, bool) {
	m := boundsRe.FindStringSubmatch(s)
	if m == nil {
		return Bounds{}, false
	}
	x1, _ := strconv.Atoi(m[1])
	y1, _ := strconv.Atoi(m[2])
	x2, _ := strconv.Atoi(m[3])
	y2, _ := strconv.Atoi(m[4])
	return Bounds{X1: x1, Y1: y1, X2: x2, Y2: y2}, true
}

// FindByText returns the first element whose text or content-description matches
// query. When partial is true it does a case-insensitive substring match;
// otherwise it requires a case-insensitive exact match. Clickable elements are
// preferred over non-clickable ones when several match.
func FindByText(elems []Element, query string, partial bool) (Element, bool) {
	q := strings.ToLower(strings.TrimSpace(query))
	var fallback *Element
	for i := range elems {
		e := &elems[i]
		if matches(e.Text, q, partial) || matches(e.Desc, q, partial) {
			if e.Clickable {
				return *e, true
			}
			if fallback == nil {
				fallback = e
			}
		}
	}
	if fallback != nil {
		return *fallback, true
	}
	return Element{}, false
}

func matches(field, q string, partial bool) bool {
	f := strings.ToLower(strings.TrimSpace(field))
	if f == "" {
		return false
	}
	if partial {
		return strings.Contains(f, q)
	}
	return f == q
}
