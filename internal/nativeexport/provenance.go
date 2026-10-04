package nativeexport

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

const provenanceSchema = "pptxgengo.local-native-issuance.v1"

// Provenance binds the complete receipt to a locally trusted render issuer.
// It is local issuance evidence, not OS attestation: code running as this user
// can access the private key. A receipt cannot nominate its own trusted key.
type Provenance struct {
	Schema    string `json:"schema"`
	KeyID     string `json:"key_id"`
	IssuedAt  string `json:"issued_at"`
	Signature string `json:"signature"`
}

type issuerKey struct {
	Seed string `json:"seed"`
}

func issuerKeyPath() (string, error) {
	root, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "pptxgengo", "native-trust", "issuer-key.json"), nil
}

func readIssuerKey(path string) (ed25519.PrivateKey, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Mode().Perm()&0077 != 0 {
		return nil, fmt.Errorf("native issuer key must be a private regular file (mode 0600): %s", path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var stored issuerKey
	if err = strictReceiptJSON(data, &stored); err != nil {
		return nil, fmt.Errorf("invalid native issuer key: %w", err)
	}
	seed, err := base64.StdEncoding.DecodeString(stored.Seed)
	if err != nil || len(seed) != ed25519.SeedSize {
		return nil, fmt.Errorf("invalid native issuer seed")
	}
	return ed25519.NewKeyFromSeed(seed), nil
}

func localIssuerKey(create bool) (ed25519.PrivateKey, error) {
	path, err := issuerKeyPath()
	if err != nil {
		return nil, err
	}
	key, err := readIssuerKey(path)
	if err == nil || !os.IsNotExist(err) || !create {
		return key, err
	}
	dir := filepath.Dir(path)
	if err = os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}
	info, err := os.Lstat(dir)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() || info.Mode().Perm()&0077 != 0 {
		return nil, fmt.Errorf("native trust directory must be private (mode 0700): %s", dir)
	}
	_, key, err = ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(issuerKey{Seed: base64.StdEncoding.EncodeToString(key.Seed())})
	if err != nil {
		return nil, err
	}
	// Publish a complete key exclusively. Concurrent renders share the winner;
	// readers cannot observe a partially written seed or an overwritten key.
	f, err := os.CreateTemp(dir, ".issuer-key-")
	if err != nil {
		return nil, err
	}
	defer os.Remove(f.Name())
	if _, err = f.Write(data); err != nil {
		f.Close()
		return nil, err
	}
	if err = f.Close(); err != nil {
		return nil, err
	}
	if err = os.Link(f.Name(), path); err != nil && !os.IsExist(err) {
		return nil, err
	}
	return readIssuerKey(path)
}

func receiptSigningBytes(receipt Receipt) ([]byte, error) {
	if receipt.Provenance == nil {
		return nil, fmt.Errorf("unsigned native receipt; rerun render with this CLI")
	}
	p := *receipt.Provenance
	p.Signature = ""
	receipt.Provenance = &p
	return json.Marshal(receipt)
}

// Only the successful render path calls the issuer. There is deliberately no
// exported signing API or CLI command for signing a supplied receipt.
func signReceipt(receipt *Receipt) error {
	key, err := localIssuerKey(true)
	if err != nil {
		return fmt.Errorf("native provenance issuance failed: %w", err)
	}
	receipt.Provenance = &Provenance{Schema: provenanceSchema, KeyID: hash(key.Public().(ed25519.PublicKey)), IssuedAt: time.Now().UTC().Format(time.RFC3339Nano)}
	data, err := receiptSigningBytes(*receipt)
	if err != nil {
		return err
	}
	receipt.Provenance.Signature = base64.StdEncoding.EncodeToString(ed25519.Sign(key, data))
	return nil
}

func strictReceiptJSON(data []byte, value any) error {
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if err := d.Decode(value); err != nil {
		return err
	}
	var extra any
	if err := d.Decode(&extra); err != io.EOF {
		return fmt.Errorf("expected exactly one JSON document")
	}
	return nil
}

// VerifyReceipt verifies local issuance before a receipt can establish native
// coverage. Artifact hashes and source/build identity still require checking by
// the caller. Verification never creates a key or trusts a receipt-supplied key.
func VerifyReceipt(data []byte) (Receipt, error) {
	var receipt Receipt
	if err := strictReceiptJSON(data, &receipt); err != nil {
		return receipt, fmt.Errorf("invalid native receipt: %w", err)
	}
	if receipt.Provenance == nil || receipt.Provenance.Schema != provenanceSchema {
		return receipt, fmt.Errorf("native receipt has no trusted local issuance signature; rerun render with this CLI")
	}
	if _, err := time.Parse(time.RFC3339Nano, receipt.Provenance.IssuedAt); err != nil {
		return receipt, fmt.Errorf("invalid native issuance time")
	}
	key, err := localIssuerKey(false)
	if err != nil {
		return receipt, fmt.Errorf("native issuer is not trusted by this local installation: %w", err)
	}
	public := key.Public().(ed25519.PublicKey)
	if receipt.Provenance.KeyID != hash(public) {
		return receipt, fmt.Errorf("native receipt issuer is not the locally trusted key")
	}
	signature, err := base64.StdEncoding.DecodeString(receipt.Provenance.Signature)
	if err != nil || len(signature) != ed25519.SignatureSize {
		return receipt, fmt.Errorf("invalid native receipt signature")
	}
	payload, err := receiptSigningBytes(receipt)
	if err != nil || !ed25519.Verify(public, payload, signature) {
		return receipt, fmt.Errorf("native receipt signature verification failed")
	}
	return receipt, nil
}
