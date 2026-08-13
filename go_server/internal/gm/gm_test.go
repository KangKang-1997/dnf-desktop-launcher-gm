package gm

import (
	"encoding/hex"
	"testing"
)

func TestDecodeNameHexRepairsMojibake(t *testing.T) {
	want := "\u52c7\u58eb"
	mojibake := "\u00e5\u2039\u2021\u00e5\u00a3\u00ab"
	got := decodeNameHex(hex.EncodeToString([]byte(mojibake)))
	if got != want {
		t.Fatalf("decodeNameHex()=%q, want %q", got, want)
	}
}

func TestDecodeNameHexASCII(t *testing.T) {
	got := decodeNameHex(hex.EncodeToString([]byte("Hero123")))
	if got != "Hero123" {
		t.Fatalf("decodeNameHex()=%q, want Hero123", got)
	}
}
