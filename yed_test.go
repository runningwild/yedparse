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
    for _,edge := range g.Graph.Edges {
      if edge.R == 255 {
        red_count++
      }
      if edge.G == 255 {
        green_count++
      }
      if edge.R == 255 && edge.G == 255 {
        panic("Shoudn't have found an edge that has both R and G set to 255.")
      }
    }
    c.Expect(red_count, Equals, 2)
    c.Expect(green_count, Equals, 2)
  })
}