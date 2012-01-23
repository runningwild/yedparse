package yed

import (
  "fmt"
  "encoding/xml"
  "os"
  "strconv"
  "io"
  "io/ioutil"
  "bytes"
  "strings"
)

type attribute struct {
  Key  string `xml:"key,attr"`
  Type string `xml:"type,attr"`
  Data string `xml:",chardata"`
}
func (a *attribute) Int() int {
  if a.Type != "int" {
    panic(fmt.Sprintf("Tried to get an attribute of type %s as an int.", a.Type))
  }
  v,err := strconv.Atoi(a.Data)
  if err != nil {
    panic(err.Error())
  }
  return v
}
func (a *attribute) Float64() float64 {
  if a.Type != "double" {
    panic(fmt.Sprintf("Tried to get an attribute of type %s as a double.", a.Type))
  }
  v,err := strconv.ParseFloat(a.Data, 64)
  if err != nil {
    panic(err.Error())
  }
  return v
}
func (a *attribute) Str() string {
  if a.Type != "String" {
    panic(fmt.Sprintf("Tried to get an attribute of type %s as a string.", a.Type))
  }
  return a.Data
}

type Section struct {
  Name       string      `xml:"name,attr"`
  Attributes []attribute `xml:"attribute"`
  Sections   []Section   `xml:"section"`

  atts map[string]*attribute
}
func (s *Section) GetAttribute(name string) *attribute {
  if s.atts == nil {
    s.atts = make(map[string]*attribute)
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
  g.hierarchic = s.GetAttribute("hierarchic").Int()
  g.label = s.GetAttribute("label").Str()
  g.directed = s.GetAttribute("directed").Int()
  g.nodes = make(map[int]*Node)
  for _,section := range s.Sections {
    if section.Name != "node" { continue }
    node,err := section.MakeNode(&g)
    if err != nil {
      return nil, err
    }
    g.nodes[node.id] = node
  }
  for _,node := range g.nodes {
    if node.group_id >= 0 {
      kids := g.nodes[node.group_id].children
      kids = append(kids, node)
      g.nodes[node.group_id].children = kids
    }
  }
  for _,section := range s.Sections {
    if section.Name != "edge" { continue }
    edge,err := section.MakeEdge(&g)
    if err != nil {
      return nil, err
    }
    g.edges = append(g.edges, edge)
    src := g.nodes[edge.src]
    src.outputs = append(src.outputs, edge)
    dst := g.nodes[edge.dst]
    dst.inputs = append(dst.inputs, edge)
  }
  return &g, nil
}
func (s *Section) MakeNode(graph *Graph) (*Node, error) {
  if s.Name != "node" {
    return nil, &Error{ "Nodes can only be made out of 'node' sections." }
  }
  var n Node
  n.graph = graph
  n.id = s.GetAttribute("id").Int()
  n.label = s.GetAttribute("label").Str()
  n.is_group = (s.GetAttribute("isGroup") != nil)
  att := s.GetAttribute("gid")
  if att == nil {
    n.group_id = -1
  } else {
    n.group_id = att.Int()
  }
  n.process()
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
func (s *Section) MakeEdge(graph *Graph) (*Edge, error) {
  if s.Name != "edge" {
    return nil, &Error{ "Edges can only be made out of 'edge' sections." }
  }
  var e Edge
  e.graph = graph
  e.src = s.GetAttribute("source").Int()
  e.dst = s.GetAttribute("target").Int()
  label := s.GetAttribute("label")
  if label != nil {
    e.label = label.Str()
  }
  for _,sec := range s.Sections {
    switch sec.Name {
      case "graphics":
        fill := sec.GetAttribute("fill")
        if fill == nil { continue }
        s := fill.Str()
        if len(s) != 7 { continue }
        e.r = hexToInt(s[1:3])
        e.g = hexToInt(s[3:5])
        e.b = hexToInt(s[5:7])
    }
  }
  e.process()
  return &e, nil
}

type Document struct {
  Creator string
  Version string
  Graph   Graph
}
type Graph struct {
  hierarchic int
  label      string
  directed   int
  nodes      map[int]*Node
  edges      []*Edge
}
func (g *Graph) NumEdges() int {
  return len(g.edges)
}
func (g *Graph) Edge(n int) *Edge {
  return g.edges[n]
}
func (g *Graph) NumNodes() int {
  return len(g.nodes)
}
func (g *Graph) Node(n int) *Node {
  return g.nodes[n]
}
type labeler struct {
  // The text associated with this node in the yed file
  label   string

  // The label text split into lines
  lines []string

  // For any line in lines that is of the form "foo:bar" this map will contain a key 'foo'
  // with the value 'bar'
  tags map[string]string
}
func (l *labeler) process() {
  l.lines = strings.Split(l.label, "\n")
  l.tags = make(map[string]string)
  for _,line := range l.lines {
    if strings.Contains(line, ":") {
      parts := strings.SplitN(line, ":", 2)
      l.tags[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
    }
  }
}
func (l *labeler) Label() string {
  return l.label
}
func (l *labeler) NumLines() int {
  return len(l.label)
}
func (l *labeler) Line(n int) string {
  return l.lines[n]
}
func (l *labeler) Tag(key string) string {
  return l.tags[key]
}
type Node struct {
  graph *Graph

  id      int

  labeler

  group_id int
  is_group bool

  inputs  []*Edge
  outputs []*Edge
  children []*Node
}
func (n *Node) NumInputs() int {
  return len(n.inputs)
}
func (n *Node) Input(id int) *Edge {
  return n.inputs[id]
}
func (n *Node) NumOutputs() int {
  return len(n.outputs)
}
func (n *Node) Output(id int) *Edge {
  return n.outputs[id]
}
func (n *Node) NumChildren() int {
  return len(n.children)
}
func (n *Node) Child(id int) *Node {
  return n.children[id]
}
// Returns the Node representing the group that this Node belongs to, or nil
// if this Node doesn't belong to a group.
func (n *Node) Group() *Node {
  return n.graph.nodes[n.group_id]
}

type Edge struct {
  graph *Graph

  src   int
  dst   int
  labeler
  r,g,b int
}
func (e *Edge) Src() *Node {
  return e.graph.nodes[e.src]
}
func (e *Edge) Dst() *Node {
  return e.graph.nodes[e.dst]
}
func (e *Edge) RGBA() (r,g,b,a uint32) {
  r = uint32(e.r)
  g = uint32(e.g)
  b = uint32(e.b)
  a = 255
  return
}
func Parse(r io.Reader) (*Document, error) {
  var s Section
  data,err := ioutil.ReadAll(r)
  if err != nil {
    return nil, err
  }
  for i := 0; i < len(data) - 1; i++ {
    if data[i] == '?' && data[i+1] == '>' {
      data = data[i+2:]
      break
    }
  }
  r = bytes.NewBuffer(data)

  err = xml.Unmarshal(r, &s)
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
