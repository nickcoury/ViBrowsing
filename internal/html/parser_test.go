package html

import (
	"fmt"
	"os"
	"testing"
)

func TestParserDebug(t *testing.T) {
	data, err := os.ReadFile("../../sample_pages/test1.html")
	if err != nil {
		t.Fatal(err)
	}

	dom := Parse(data)
	fmt.Printf("DOM:\n%s\n", dom.String())
}
