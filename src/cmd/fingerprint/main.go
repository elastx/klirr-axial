package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"axial/models"
)

func readAll(r io.Reader) (string, error) {
    b := strings.Builder{}
    scanner := bufio.NewScanner(r)
    scanner.Buffer(make([]byte, 0, 1024), 10*1024*1024)
    for scanner.Scan() {
        b.WriteString(scanner.Text())
        b.WriteByte('\n')
    }
    if err := scanner.Err(); err != nil {
        return "", err
    }
    return strings.TrimSpace(b.String()), nil
}

func main() {
    // Read armored public key from stdin
    armored, err := readAll(os.Stdin)
    if err != nil || armored == "" {
        fmt.Fprintln(os.Stderr, "failed to read armored public key from stdin")
        os.Exit(1)
    }

    pk := models.PublicKey(armored)
    // Prefer encryption subkey KeyID; fallback to primary/signing key ID
    if encFps, err := pk.GetEncryptionFingerprints(); err == nil && len(encFps) > 0 {
        // Use the first encryption recipient ID deterministically
        fmt.Println(strings.ToLower(string(encFps[0])))
        return
    }

    // Fallback: primary/signing key ID
    fp, err := pk.GetFingerprint()
    if err != nil {
        fmt.Fprintln(os.Stderr, "failed to parse public key:", err)
        os.Exit(2)
    }
    fmt.Println(strings.ToLower(string(fp)))
}
