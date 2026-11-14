package goquery

import (
	"testing"

	"github.com/andybalholm/cascadia"
)

func TestAdd(t *testing.T) {
	sel := Doc().Find("div.row-fluid").Add("a")
	assertLength(t, sel.Nodes, 19)
}

func TestAddInvalid(t *testing.T) {
	sel1 := Doc().Find("div.row-fluid")
	sel2 := sel1.Add("")
	assertLength(t, sel1.Nodes, 9)
	assertLength(t, sel2.Nodes, 9)
	if sel1 == sel2 {
		t.Errorf("selections should not be the same")
	}
}

func TestAddRollback(t *testing.T) {
	sel := Doc().Find(".pvk-content")
	sel2 := sel.Add("a").End()
	assertEqual(t, sel, sel2)
}

func TestAddSelection(t *testing.T) {
	sel := Doc().Find("div.row-fluid")
	sel2 := Doc().Find("a")
	sel = sel.AddSelection(sel2)
	assertLength(t, sel.Nodes, 19)
}

func TestAddSelectionNil(t *testing.T) {
	sel := Doc().Find("div.row-fluid")
	assertLength(t, sel.Nodes, 9)

	sel = sel.AddSelection(nil)
	assertLength(t, sel.Nodes, 9)
}

func TestAddSelectionRollback(t *testing.T) {
	sel := Doc().Find(".pvk-content")
	sel2 := sel.Find("a")
	sel2 = sel.AddSelection(sel2).End()
	assertEqual(t, sel, sel2)
}

func TestAddNodes(t *testing.T) {
	sel := Doc().Find("div.pvk-gutter")
	sel2 := Doc().Find(".pvk-content")
	sel = sel.AddNodes(sel2.Nodes...)
	assertLength(t, sel.Nodes, 9)
}

func TestAddNodesNone(t *testing.T) {
	sel := Doc().Find("div.pvk-gutter").AddNodes()
	assertLength(t, sel.Nodes, 6)
}

func TestAddNodesRollback(t *testing.T) {
	sel := Doc().Find(".pvk-content")
	sel2 := sel.Find("a")
	sel2 = sel.AddNodes(sel2.Nodes...).End()
	assertEqual(t, sel, sel2)
}

func TestAddNodesBig(t *testing.T) {
	doc := DocW()
	sel := doc.Find("li")
	assertLength(t, sel.Nodes, 373)
	sel2 := doc.Find("xyz")
	assertLength(t, sel2.Nodes, 0)

	nodes := sel.Nodes
	sel2 = sel2.AddNodes(nodes...)
	assertLength(t, sel2.Nodes, 373)
	nodes2 := append(nodes, nodes...)
	sel2 = sel2.End().AddNodes(nodes2...)
	assertLength(t, sel2.Nodes, 373)
	nodes3 := append(nodes2, nodes...)
	sel2 = sel2.End().AddNodes(nodes3...)
	assertLength(t, sel2.Nodes, 373)
}

func TestAndSelf(t *testing.T) {
	sel := Doc().Find(".span12").Last().AndSelf()
	assertLength(t, sel.Nodes, 2)
}

func TestAndSelfRollback(t *testing.T) {
	sel := Doc().Find(".pvk-content")
	sel2 := sel.Find("a").AndSelf().End().End()
	assertEqual(t, sel, sel2)
}

func TestAddBack(t *testing.T) {
	sel := Doc().Find(".span12").Last().AddBack()
	assertLength(t, sel.Nodes, 2)
}

func TestAddBackRollback(t *testing.T) {
	sel := Doc().Find(".pvk-content")
	sel2 := sel.Find("a").AddBack().End().End()
	assertEqual(t, sel, sel2)
}

func TestAddBackFiltered(t *testing.T) {
	sel := Doc().Find(".span12, .footer").Find("h1").AddBackFiltered(".footer")
	assertLength(t, sel.Nodes, 2)
}

func TestAddBackFilteredRollback(t *testing.T) {
	sel := Doc().Find(".span12, .footer")
	sel2 := sel.Find("h1").AddBackFiltered(".footer").End().End()
	assertEqual(t, sel, sel2)
}

// TestAddBackNilPrevSelMatrix covers the empty-fallback cases where there is
// no previous selection on the stack: a chain starting at the document root,
// a Selection produced by Clone, and an End() that falls back past the chain
// head. In every such scenario the filtered forms must converge to the same
// safe result as the unfiltered AddBack - return the current set instead of
// panicking on the nil previous selection.
func TestAddBackNilPrevSelMatrix(t *testing.T) {
	scenarios := []struct {
		name string
		sel  func() *Selection
	}{
		{
			name: "document root",
			sel:  func() *Selection { return Doc().Selection },
		},
		{
			name: "cloned selection",
			sel:  func() *Selection { return Doc().Find(".pvk-content").Clone() },
		},
		{
			name: "end past chain head",
			sel:  func() *Selection { return Doc().Find(".pvk-content").End() },
		},
	}

	forms := []struct {
		name string
		fn   func(*Selection) *Selection
	}{
		{
			name: "unfiltered",
			fn:   func(s *Selection) *Selection { return s.AddBack() },
		},
		{
			name: "filtered",
			fn:   func(s *Selection) *Selection { return s.AddBackFiltered("div") },
		},
		{
			name: "matcher",
			fn: func(s *Selection) *Selection {
				return s.AddBackMatcher(cascadia.MustCompile("div"))
			},
		},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			for _, form := range forms {
				t.Run(form.name, func(t *testing.T) {
					src := sc.sel()
					if src.prevSel != nil {
						t.Fatalf("%s: expected nil prevSel, got %v", sc.name, src.prevSel)
					}

					// None of the forms may panic on the empty fallback.
					got := form.fn(src)

					// The empty fallback must keep the current set verbatim:
					// same length, same node pointers in the same order.
					if len(got.Nodes) != len(src.Nodes) {
						t.Fatalf("expected %d nodes, got %d", len(src.Nodes), len(got.Nodes))
					}
					for i, n := range got.Nodes {
						if n != src.Nodes[i] {
							t.Fatalf("node %d: expected current set to be returned verbatim, got a different node", i)
						}
					}

					// Filtered and unfiltered forms must produce the same set.
					want := src.AddBack()
					if len(got.Nodes) != len(want.Nodes) {
						t.Fatalf("filtered form returned %d nodes, unfiltered form %d", len(got.Nodes), len(want.Nodes))
					}
					for i := range got.Nodes {
						if got.Nodes[i] != want.Nodes[i] {
							t.Fatalf("node %d: filtered form diverges from unfiltered AddBack", i)
						}
					}
				})
			}
		})
	}
}

// TestAddBackNilPrevSelEmptySet checks the empty current set variant of the
// nil-prevSel fallback: an empty cloned selection must also return safely.
func TestAddBackNilPrevSelEmptySet(t *testing.T) {
	src := Doc().Find(".pvk-zzcontentzz").Clone()
	if src.prevSel != nil {
		t.Fatalf("expected nil prevSel, got %v", src.prevSel)
	}
	assertLength(t, src.AddBack().Nodes, 0)
	assertLength(t, src.AddBackFiltered("div").Nodes, 0)
	assertLength(t, src.AddBackMatcher(cascadia.MustCompile("div")).Nodes, 0)
}
