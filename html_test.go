package apinovo

import (
	"strings"
	"testing"
)

func TestTextToHTML(t *testing.T) {
	doc := `
	<html>
	<title>hello</title>
	<body>world<a>africa</a></body>
	</html>
	`
	rst := TextFromHTML(strings.NewReader(doc))
	sample := []string{"hello", "world", "africa"}
	if len(rst) != len(sample) {
		t.Fatalf("expected %d got %d", len(sample), len(rst))
	}

	for k := range rst {
		if rst[k] != sample[k] {
			t.Errorf("expected %s got %s", sample[k], rst[k])
		}
	}

}
