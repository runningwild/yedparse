package yed_test

import (
  "github.com/orfjackal/gospec/src/gospec"
  "testing"
)


func TestAllSpecs(t *testing.T) {
  r := gospec.NewRunner()
  r.AddSpec(YedSpec)
  gospec.MainGoTest(r, t)
}

