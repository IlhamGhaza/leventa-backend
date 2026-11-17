package usecase

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// NormalizeWhatsApp converts an Indonesian-style number to digits with the
// 62 country code (e.g. "0812-345 678" -> "62812345678"). ok is false when the
// result is not a plausible international number.
func NormalizeWhatsApp(raw string) (string, bool) {
	var b strings.Builder
	for _, r := range raw {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	d := b.String()
	switch {
	case strings.HasPrefix(d, "0"):
		d = "62" + d[1:]
	case strings.HasPrefix(d, "8"):
		d = "62" + d
	}
	if len(d) < 9 || len(d) > 15 || !strings.HasPrefix(d, "62") {
		return "", false
	}
	return d, true
}

// BuildWhatsAppURL builds a wa.me link with a URL-encoded prefilled message.
// number must be the representing marketing user's number, never the owner's.
func BuildWhatsAppURL(number, title, code string, price *float64, location string) string {
	msg := fmt.Sprintf("Hello, I am interested in %s (%s)", title, code)
	if price != nil {
		msg += ", listed at " + formatRupiah(*price)
	}
	if location != "" {
		msg += " in " + location
	}
	msg += ". Is this property still available?"
	enc := strings.ReplaceAll(url.QueryEscape(msg), "+", "%20")
	return "https://wa.me/" + number + "?text=" + enc
}

// formatRupiah renders 2500000000 as "Rp 2.500.000.000".
func formatRupiah(v float64) string {
	s := strconv.FormatInt(int64(v+0.5), 10)
	var out []byte
	for i, c := range []byte(s) {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, '.')
		}
		out = append(out, c)
	}
	return "Rp " + string(out)
}
