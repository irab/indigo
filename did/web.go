package did

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"unicode"

	"github.com/whyrusleeping/go-did"
)

type WebResolver struct {
	// TODO: cache? maybe at a different layer
}

func (wr *WebResolver) GetDocument(ctx context.Context, didstr string) (*Document, error) {
	pdid, err := did.ParseDID(didstr)
	if err != nil {
		return nil, err
	}

	val := pdid.Value()
	if err := checkValidDidWeb(val); err != nil {
		return nil, err
	}

	resp, err := http.Get("https://" + val + "/.well-known/did.json")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("fetch did request failed (status %d): %s", resp.StatusCode, resp.Status)
	}

	var out did.Document
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}

	return &out, nil
}

var disallowedTlds = map[string]bool{
	"example":  true,
	"invalid":  true,
	"local":    true,
	"arpa":     true,
	"onion":    true,
	"internal": true,
}

func checkValidDidWeb(val string) error {
	// no ports or ipv6
	if strings.Contains(val, ":") {
		return fmt.Errorf("did:web resolver does not handle ports or documents at sub-paths")
	}
	// no trailing '.'
	if strings.HasSuffix(val, ".") {
		return fmt.Errorf("cannot have trailing period in hostname")
	}

	parts := strings.Split(val, ".")
	if len(parts) == 1 {
		return fmt.Errorf("no bare hostnames (must have subdomain)")
	}

	tld := parts[len(parts)-1]
	if disallowedTlds[tld] {
		return fmt.Errorf("domain cannot use any disallowed TLD")
	}

	// disallow tlds that start with numbers
	if unicode.IsNumber(rune(tld[0])) {
		return fmt.Errorf("TLD cannot start with a number")
	}

	return nil
}
