// Package nativepkg implements a source-derived native OOXML scene experiment.
// It deliberately preserves PowerPoint's full object/style vocabulary rather
// than pretending an arbitrary imported slide is a high-level layout contract.
package nativepkg

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

type Attr struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
type Node struct {
	Name     string  `json:"name,omitempty"`
	Attrs    []Attr  `json:"attributes,omitempty"`
	Text     string  `json:"text,omitempty"`
	Children []*Node `json:"children,omitempty"`
}

func qname(n xml.Name) string {
	if n.Space != "" {
		return n.Space + ":" + n.Local
	}
	return n.Local
}
func local(s string) string {
	if i := strings.IndexByte(s, ':'); i >= 0 {
		return s[i+1:]
	}
	return s
}
func Parse(b []byte) (*Node, error) {
	d := xml.NewDecoder(bytes.NewReader(b))
	var root *Node
	var stack []*Node
	for {
		t, e := d.RawToken()
		if e == io.EOF {
			break
		}
		if e != nil {
			return nil, e
		}
		switch v := t.(type) {
		case xml.StartElement:
			n := &Node{Name: qname(v.Name)}
			for _, a := range v.Attr {
				n.Attrs = append(n.Attrs, Attr{qname(a.Name), a.Value})
			}
			if len(stack) > 0 {
				p := stack[len(stack)-1]
				p.Children = append(p.Children, n)
			} else {
				root = n
			}
			stack = append(stack, n)
		case xml.EndElement:
			if len(stack) == 0 {
				return nil, fmt.Errorf("unbalanced XML")
			}
			stack = stack[:len(stack)-1]
		case xml.CharData:
			if len(stack) > 0 {
				p := stack[len(stack)-1]
				p.Children = append(p.Children, &Node{Text: string(v)})
			}
		}
	}
	if root == nil || len(stack) != 0 {
		return nil, fmt.Errorf("incomplete XML")
	}
	return root, nil
}
func (n *Node) Attr(k string) string {
	for _, a := range n.Attrs {
		if a.Name == k {
			return a.Value
		}
	}
	return ""
}
func (n *Node) SetAttr(k, v string) {
	for i := range n.Attrs {
		if n.Attrs[i].Name == k {
			n.Attrs[i].Value = v
			return
		}
	}
	n.Attrs = append(n.Attrs, Attr{k, v})
}
func (n *Node) Child(k string) *Node {
	for _, c := range n.Children {
		if local(c.Name) == k {
			return c
		}
	}
	return nil
}
func (n *Node) Walk(f func(*Node)) {
	f(n)
	for _, c := range n.Children {
		c.Walk(f)
	}
}
func (n *Node) TextContent() string {
	var b strings.Builder
	n.Walk(func(c *Node) {
		if c.Name == "" {
			b.WriteString(c.Text)
		}
	})
	return b.String()
}
func escape(b *bytes.Buffer, s string) { _ = xml.EscapeText(b, []byte(s)) }
func (n *Node) write(b *bytes.Buffer) {
	if n.Name == "" {
		escape(b, n.Text)
		return
	}
	b.WriteByte('<')
	b.WriteString(n.Name)
	for _, a := range n.Attrs {
		b.WriteByte(' ')
		b.WriteString(a.Name)
		b.WriteString(`="`)
		escape(b, a.Value)
		b.WriteByte('"')
	}
	if len(n.Children) == 0 {
		b.WriteString("/>")
		return
	}
	b.WriteByte('>')
	for _, c := range n.Children {
		c.write(b)
	}
	b.WriteString("</")
	b.WriteString(n.Name)
	b.WriteByte('>')
}
func (n *Node) XML() []byte {
	var b bytes.Buffer
	b.WriteString(xml.Header)
	n.write(&b)
	return b.Bytes()
}
