package pvf

import "testing"

func TestDecryptBytesRoundTripRotation(t *testing.T) {
	input := []byte{0x11, 0x90, 0xA7, 0x81}
	got := decryptBytes(input, 0)
	want := []byte{0x00, 0x00, 0x00, 0x00}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("decryptBytes()=%v, want %v", got, want)
		}
	}
}

func TestListToDictSegment(t *testing.T) {
	types := []int{5, 7, 7, 5, 7, 5}
	values := []any{"[name]", "a", "b", "[growtype name]", "fighter", "[/growtype name]"}
	got := listToDict(types, values)
	name := joinValues(listFromDetail(got["[name]"]))
	if name != "ab" {
		t.Fatalf("[name]=%q, want ab", name)
	}
	grow := joinValues(listFromDetail(got["[growtype name]"]))
	if grow != "fighter" {
		t.Fatalf("grow name=%q, want fighter", grow)
	}
}

func TestGetExpForLevelIndexing(t *testing.T) {
	table := []any{float64(100), float64(300)}
	value, ok := intFromJSON(table[0])
	if !ok || value != 100 {
		t.Fatalf("intFromJSON=%d/%v, want 100/true", value, ok)
	}
}
