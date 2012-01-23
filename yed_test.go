package yed_test

import (
  . "gospec"
  "gospec"
  "yed"
)

func YedSpec(c gospec.Context) {
  c.Specify("Load a simple .xgml file.", func() {
    g, err := yed.ParseFromFile("state.xgml")
    c.Assume(err, Equals, nil)
    red_count := 0
    green_count := 0
    for i := 0; i < g.Graph.NumEdges(); i++ {
      edge := g.Graph.Edge(i)
      r,g,_,_ := edge.RGBA()
      if r == 255 {
        red_count++
      }
      if g == 255 {
        green_count++
      }
      if r == 255 && g == 255 {
        panic("Shoudn't have found an edge that has both R and G set to 255.")
      }
    }
    c.Expect(red_count, Equals, 2)
    c.Expect(green_count, Equals, 2)

    // Check that certain nodes and edges are hooked up properly
    found := false
    for i := 0; i < g.Graph.NumNodes(); i++ {
      node := g.Graph.Node(i)
      if node.Label() == "option 1" {
        found = true
        c.Expect(node.NumInputs(), Equals, 1)
        c.Assume(node.NumOutputs(), Equals, 2)
        for i := 0; i < node.NumOutputs(); i++ {
          edge := node.Output(i)
          if edge.Line(0) == "Edge Foo" {
            c.Expect(edge.Tag("tag1"), Equals, "monkey")
            c.Expect(edge.Tag("tag2"), Equals, "chimp")
            c.Expect(edge.Dst().Label(), Equals, "option 3")
          } else if edge.Line(0) == "Edge Bar" {
            c.Expect(edge.Tag("tag3"), Equals, "walrus")
            c.Expect(edge.Dst().Label(), Equals, "option 4")
          } else {
            panic("Expected 'Edge Foo' or 'Edge Bar', found '" + edge.Line(0) + "'")
          }
        }
      }
    }

    // Check that groups are set up properly
    for i := 0; i < g.Graph.NumNodes(); i++ {
      group := g.Graph.Node(i).Group()
      if group == nil { continue }
      if group.NumLines() == 0 || group.Line(0) != "nubcake" { continue }
      c.Assume(group.NumChildren(), Equals, 2)
      var opt3,opt4 int
      for i := 0; i < group.NumChildren(); i++ {
        switch group.Child(i).Label() {
          case "option 3":
          opt3++

          case "option 4":
          opt4++

          default:
          panic("Expected 'option3' or 'option4', found '" + group.Child(i).Label() + "'")
        }
      }
      c.Expect(opt3, Equals, 1)
      c.Expect(opt4, Equals, 1)

      c.Expect(group.NumInputs(), Equals, 0)
      c.Assume(group.NumOutputs(), Equals, 1)
      c.Expect(group.Output(0).Dst().Label(), Equals, "choice")

      parent := group.Group()
      c.Assume(parent, Not(Equals), nil)
      c.Expect(parent.Line(0), Equals, "bigger")
      c.Expect(parent.Tag("tag4"), Equals, "tiger")
      c.Expect(parent.NumChildren(), Equals, 3)
    }


    c.Expect(found, Equals, true)
  })
}