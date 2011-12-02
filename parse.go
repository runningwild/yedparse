package yed

import (
  "fmt"
  "encoding/xml"
  "os"
  "strconv"
  "io"
)

type Attribute struct {
  Key  string `xml:"attr"`
  Type string `xml:"attr"`
  Data string `xml:"chardata"`
}
func (a *Attribute) Int() int {
  if a.Type != "int" {
    panic(fmt.Sprintf("Tried to get an Attribute of type %s as an int.", a.Type))
  }
  v,err := strconv.Atoi(a.Data)
  if err != nil {
    panic(err.Error())
  }
  return v
}
func (a *Attribute) Float64() float64 {
  if a.Type != "double" {
    panic(fmt.Sprintf("Tried to get an Attribute of type %s as a double.", a.Type))
  }
  v,err := strconv.Atof64(a.Data)
  if err != nil {
    panic(err.Error())
  }
  return v
}
func (a *Attribute) Str() string {
  if a.Type != "String" {
    panic(fmt.Sprintf("Tried to get an Attribute of type %s as a string.", a.Type))
  }
  return a.Data
}

type Section struct {
  Name       string      `xml:"attr"`
  Attributes []Attribute `xml:"attribute>"`
  Sections   []Section   `xml:"section>"`

  atts map[string]*Attribute
}
func (s *Section) GetAttribute(name string) *Attribute {
  if s.atts == nil {
    s.atts = make(map[string]*Attribute)
    for i,att := range s.Attributes {
      s.atts[att.Key] = &s.Attributes[i]
    }
  }
  return s.atts[name]
}

type Error struct {
  ErrorString string
}
func (e *Error) Error() string {
  return e.ErrorString
}

func (s *Section) MakeDocument() (*Document, error) {
  if s.Name != "xgml" {
    return nil, &Error{ "Documents can only be made out of 'xgml' sections." }
  }
  var doc Document
  doc.Creator = s.GetAttribute("Creator").Str()
  doc.Version = s.GetAttribute("Version").Str()
  for _,section := range s.Sections {
    if section.Name == "graph" {
      g,err := section.MakeGraph()
      if err != nil {
        return nil, err
      }
      doc.Graph = *g
    }
    break
  }
  return &doc,nil
}

func (s *Section) MakeGraph() (*Graph, error) {
  if s.Name != "graph" {
    return nil, &Error{ "Graphs can only be made out of 'graph' sections." }
  }
  var g Graph
  g.Hierarchic = s.GetAttribute("hierarchic").Int()
  g.Label = s.GetAttribute("label").Str()
  g.Directed = s.GetAttribute("directed").Int()
  g.Nodes = make(map[int]*Node)
  g.Groups = make(map[int][]*Node)
  for _,section := range s.Sections {
    if section.Name != "node" { continue }
    node,err := section.MakeNode()
    if err != nil {
      return nil, err
    }
    g.Nodes[node.Id] = node
    if node.Group >= 0 {
      groups := g.Groups[node.Group]
      groups = append(groups, node)
      g.Groups[node.Group] = groups
    }
  }
  for _,section := range s.Sections {
    if section.Name != "edge" { continue }
    edge,err := section.MakeEdge()
    if err != nil {
      return nil, err
    }
    g.Edges = append(g.Edges, edge)
    src := g.Nodes[edge.Src]
    src.Outputs = append(src.Outputs, edge)
    dst := g.Nodes[edge.Dst]
    dst.Inputs = append(dst.Inputs, edge)
  }
  return &g, nil
}
func (s *Section) MakeNode() (*Node, error) {
  if s.Name != "node" {
    return nil, &Error{ "Nodes can only be made out of 'node' sections." }
  }
  var n Node
  n.Id = s.GetAttribute("id").Int()
  n.Label = s.GetAttribute("label").Str()
  att := s.GetAttribute("gid")
  if att == nil {
    n.Group = -1
  } else {
    n.Group = att.Int()
  }
  return &n, nil
}
func hexToInt(h string) int {
  if len(h) != 2 { panic("WTF are you doing!?") }
  n := 0
  for _,c := range h {
    n *= 16
    if c >= '0' && c <= '9' {
      n += int(c - '0')
    } else if c >= 'A' && c <= 'F' {
      n += int(c - 'A') + 10
    } else {
      n += int(c - 'a') + 10
    }
  }
  return n
}
func (s *Section) MakeEdge() (*Edge, error) {
  if s.Name != "edge" {
    return nil, &Error{ "Edges can only be made out of 'edge' sections." }
  }
  var e Edge
  e.Src = s.GetAttribute("source").Int()
  e.Dst = s.GetAttribute("target").Int()
  label := s.GetAttribute("label")
  if label != nil {
    e.Label = label.Str()
  }
  for _,sec := range s.Sections {
    switch sec.Name {
      case "graphics":
        fill := sec.GetAttribute("fill")
        if fill == nil { continue }
        s := fill.Str()
        if len(s) != 7 { continue }
        e.R = hexToInt(s[1:3])
        e.G = hexToInt(s[3:5])
        e.B = hexToInt(s[5:7])
    }
  }
  return &e, nil
}

type Document struct {
  Creator string
  Version string
  Graph   Graph
}
type Graph struct {
  Hierarchic int
  Label      string
  Directed   int
  Nodes      map[int]*Node
  Edges      []*Edge
  Groups     map[int][]*Node
}
type Node struct {
  Id      int
  Label   string
  Group   int
  Inputs  []*Edge
  Outputs []*Edge
}
type Edge struct {
  Src   int
  Dst   int
  Label string
  R,G,B int
}

func Parse(r io.Reader) (*Document, error) {
  var s Section
  err := xml.Unmarshal(r, &s)
  if err != nil {
    return nil, err
  }
  doc,err := s.MakeDocument()
  if err != nil {
    fmt.Printf("Error: %s\n", err.Error())
    return nil, err
  }
  return doc, nil
}

func ParseFromFile(filename string) (*Document, error) {
  f,err := os.Open(filename)
  if err != nil {
    return nil, err
  }
  defer f.Close()
  doc,err := Parse(f)
  if err != nil {
    return nil, err
  }
  return doc, nil
}

func main() {
  doc, err := ParseFromFile("state.xgml")
  if err != nil {
    fmt.Printf("Error: %s\n", err.Error())
    return
  }
  var n *Node
  n = doc.Graph.Nodes[0]
  for i := 0; i < 50; i++ {
    fmt.Printf("%d: %s\n", i, n.Label)
    next := n.Outputs[(i*i - i) % len(n.Outputs)].Dst
    fmt.Printf("Following %d\n", next)
    n = doc.Graph.Nodes[next]
  }
  for i := range doc.Graph.Edges {
    fmt.Printf("Colors: %d %d %d\n", doc.Graph.Edges[i].R, doc.Graph.Edges[i].G, doc.Graph.Edges[i].B)
  }
}
