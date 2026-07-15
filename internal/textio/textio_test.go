package textio

import "testing"

func TestDecodeEncodePreservesLineEndingStyle(t *testing.T) {
	for _, raw := range []string{"a\nb\n", "a\r\nb\r\n"} {
		d := Decode([]byte(raw))
		if got := string(d.Encode(d.Text)); got != raw {
			t.Fatalf("%q => %q", raw, got)
		}
	}
}

func TestEncodePreservesMixedLineEndingsOutsideChanges(t *testing.T) {
	raw := "alpha\r\nbeta\ngamma\r\n"
	doc := Decode([]byte(raw))
	got := string(doc.Encode("alpha\ninserted\nbeta\ngamma\n"))
	want := "alpha\r\ninserted\r\nbeta\ngamma\r\n"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
