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
    fp, err := pk.GetFingerprint()
    if err != nil {
        fmt.Fprintln(os.Stderr, "failed to parse public key:", err)
        os.Exit(2)
    }

    // Print backend fingerprint (hex key ID)
    fmt.Println(string(fp))
}
