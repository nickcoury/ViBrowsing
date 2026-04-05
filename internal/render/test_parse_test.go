package render

import (
	"testing"
	"github.com/nickcoury/ViBrowsing/internal/css"
)

func TestParseInlineBody(t *testing.T) {
	style := "background:#eee;margin:0;padding:0"
	decls := css.ParseInline(style)
	t.Logf("Parsed %d declarations:", len(decls))
	for _, d := range decls {
		t.Logf("  %s: %s", d.Property, d.Value)
	}
	
	styleMap := css.ComputeStyle("body", "", "", decls, nil)
	t.Logf("\nComputed style for body:")
	for k, v := range styleMap {
		t.Logf("  %s: %s", k, v)
	}
}
