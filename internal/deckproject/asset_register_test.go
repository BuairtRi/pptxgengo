package deckproject

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestRegisterAssetPreservesOriginalAndRejectsOverwrite(t *testing.T) {
	p := example(t)
	original, err := os.ReadFile(filepath.Join(p.Root, "assets/originals/sample.png"))
	if err != nil {
		t.Fatal(err)
	}
	r, err := RegisterAsset(p, AssetRegistration{ID: "client-image", Data: original, Description: "Client-supplied reference image", Focus: &AssetFocus{X: .2, Y: .8}})
	if err != nil {
		t.Fatal(err)
	}
	p, err = Load(p.Root)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(p.Root, r.Path))
	if err != nil || !bytes.Equal(raw, original) {
		t.Fatal("original modified", err)
	}
	a := p.Document.Assets["client-image"]
	if a.SHA256 != digest(original) || a.Focus == nil || a.Focus.X != .2 {
		t.Fatal(a)
	}
	before := p.SourceHash()
	if _, err = RegisterAsset(p, AssetRegistration{ID: "client-image", Data: original, Description: "Another"}); err == nil {
		t.Fatal("duplicate registration")
	}
	if _, err = RegisterAsset(p, AssetRegistration{ID: "../escape", Data: original, Description: "Another"}); err == nil {
		t.Fatal("unsafe ID")
	}
	if _, err = RegisterAsset(p, AssetRegistration{ID: "bad-focus", Data: original, Description: "Another", Focus: &AssetFocus{X: 2, Y: 0}}); err == nil {
		t.Fatal("bad focus")
	}
	if _, err = RegisterAsset(p, AssetRegistration{ID: "invalid-image", Data: []byte("not an image"), Description: "Another"}); err == nil {
		t.Fatal("invalid image")
	}
	p, _ = Load(p.Root)
	if p.SourceHash() != before {
		t.Fatal("failed registrations changed source")
	}
	if _, err = os.Stat(filepath.Join(p.Root, r.Decision)); err != nil {
		t.Fatal("missing receipt", err)
	}
}
