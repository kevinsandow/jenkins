package jenkins

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/antchfx/xquery/xml"
)

// Sets the value of an attribute. Attributes are unique, will override if it already exists.
func SetAttr(n *xmlquery.Node, key, val string) {
	var attr xml.Attr
	if i := strings.Index(key, ":"); i > 0 {
		attr = xml.Attr{
			Name:  xml.Name{Space: key[:i], Local: key[i+1:]},
			Value: val,
		}
	} else {
		attr = xml.Attr{
			Name:  xml.Name{Local: key},
			Value: val,
		}
	}

	for i := range n.Attr {
		if n.Attr[i].Name.Local == attr.Name.Local && n.Attr[i].Name.Space == attr.Name.Space {
			n.Attr[i].Value = attr.Value
			return
		}
	}

	n.Attr = append(n.Attr, attr)
}

// Sets the innerText of an element, only creates a new one if none exists.
func SetElementTextNode(n *xmlquery.Node, name, val string) {
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		if child.Data == name && child.Type == xmlquery.ElementNode {
			if child.FirstChild == nil {
				addText(child, val)
				return
			}

			if child.FirstChild == child.LastChild && child.FirstChild.Type == xmlquery.TextNode {
				child.FirstChild.Data = val
				return
			}
		}
	}

	addText(AddElement(n, name), val)
}

// Appends a new element to the list of children - may produce duplicates.
func AddElement(parent *xmlquery.Node, name string) *xmlquery.Node {
	n := &xmlquery.Node{
		Data: name,
		Type: xmlquery.ElementNode,
	}

	n.Parent = parent
	if parent.FirstChild == nil {
		parent.FirstChild = n
	} else {
		parent.LastChild.NextSibling = n
		n.PrevSibling = parent.LastChild
	}

	parent.LastChild = n

	return n
}

func addText(n *xmlquery.Node, val string) {
	textNode := &xmlquery.Node{
		Data: val,
		Type: xmlquery.TextNode,
	}

	textNode.Parent = n
	if n.FirstChild == nil {
		n.FirstChild = textNode
	} else {
		n.LastChild.NextSibling = textNode
		textNode.PrevSibling = n.LastChild
	}

	n.LastChild = textNode
}

func getReaderFromNode(node *xmlquery.Node) io.Reader {
	var buf bytes.Buffer
	buf.WriteString("<?xml version='1.0' encoding='utf-8' ?>")
	for n := node.FirstChild; n != nil; n = n.NextSibling {
		outputXML(&buf, n, 0)
	}
	return &buf
}

func outputXML(buf *bytes.Buffer, n *xmlquery.Node, level int) {
	if n.Type == xmlquery.TextNode || n.Type == xmlquery.CommentNode {
		xml.EscapeText(buf, []byte(strings.TrimSpace(n.Data)))
		return
	}
	if n.Type == xmlquery.DeclarationNode {
		buf.WriteString("<?" + n.Data)
	} else {
		buf.WriteString(fmt.Sprintf("\n%s<%s", strings.Repeat("  ", level), n.Data))
	}

	for _, attr := range n.Attr {
		if attr.Name.Space != "" {
			buf.WriteString(fmt.Sprintf(` %s:%s="%s"`, attr.Name.Space, attr.Name.Local, attr.Value))
		} else {
			buf.WriteString(fmt.Sprintf(` %s="%s"`, attr.Name.Local, attr.Value))
		}
	}
	if n.Type == xmlquery.DeclarationNode {
		buf.WriteString("?>")
	} else {
		if n.FirstChild == nil {
			buf.WriteString("/>")
		} else {
			buf.WriteString(">")

			for child := n.FirstChild; child != nil; child = child.NextSibling {
				outputXML(buf, child, level+1)
			}

			if n.FirstChild == n.LastChild && n.FirstChild.Type == xmlquery.TextNode {
				buf.WriteString(fmt.Sprintf("</%s>", n.Data))
			} else {
				buf.WriteString(fmt.Sprintf("\n%s</%s>", strings.Repeat("  ", level), n.Data))
			}
		}
	}
}
