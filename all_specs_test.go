package yed_test

import (
  "gospec"
  "testing"
)


func TestAllSpecs(t *testing.T) {
  r := gospec.NewRunner()
  r.AddSpec(YedSpec)
  gospec.MainGoTest(r, t)
}

