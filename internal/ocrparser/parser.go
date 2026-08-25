package ocrparser

import (
	"math"
	"regexp"
	"strings"
)

type ParsedItem struct {
	Name       string
	Quantity   int
	UnitPrice  int64
	TotalPrice int64
	Confidence float64
}

type Parsed struct {
	MerchantName  string
	Items         []ParsedItem
	Subtotal      int64
	Tax           int64
	ServiceCharge int64
	Discount      int64
	Total         int64
}

var (
	numberToken   = regexp.MustCompile(`\d[\d.,]*`)
	qtyXUnit      = regexp.MustCompile(`(\d{1,3})\s*[@xX]\s*(\d[\d.,]*)`)
	wsRe          = regexp.MustCompile(`\s+`)
	keywordSubtot = regexp.MustCompile(`(?i)^subtotal|^sub\s+total`)
	keywordTax    = regexp.MustCompile(`(?i)^(pajak|ppn|tax|vat)\b`)
	keywordSvc    = regexp.MustCompile(`(?i)^(service|servis|svc|service\s*charge)\b`)
	keywordDisc   = regexp.MustCompile(`(?i)^(discount|diskon|disc)\b`)
	keywordTotal  = regexp.MustCompile(`(?i)^total\b`)
)

// Parse converts raw OCR text into a structured receipt.
func Parse(raw string) Parsed {
	var out Parsed
	lines := splitLines(raw)
	merchantSet := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || looksLikeDecoration(line) {
			continue
		}

		lower := strings.ToLower(line)
		switch {
		case keywordSubtot.MatchString(lower):
			out.Subtotal = lastNumber(line)
			continue
		case keywordTax.MatchString(lower):
			out.Tax = lastNumber(line)
			continue
		case keywordSvc.MatchString(lower):
			out.ServiceCharge = lastNumber(line)
			continue
		case keywordDisc.MatchString(lower):
			out.Discount = lastNumber(line)
			continue
		case keywordTotal.MatchString(lower):
			out.Total = lastNumber(line)
			continue
		}

		item, ok := parseItemLine(line)
		if ok {
			out.Items = append(out.Items, item)
			continue
		}

		if !merchantSet && !containsDigit(line) && len(line) >= 3 {
			out.MerchantName = wsRe.ReplaceAllString(line, " ")
			merchantSet = true
		}
	}

	return out
}

func splitLines(raw string) []string {
	raw = strings.ReplaceAll(raw, "\r\n", "\n")
	return strings.Split(raw, "\n")
}

func looksLikeDecoration(line string) bool {
	stripped := strings.Map(func(r rune) rune {
		switch r {
		case '-', '=', '_', '.', '*', '~', '|':
			return -1
		}
		return r
	}, line)
	return strings.TrimSpace(stripped) == ""
}

func containsDigit(s string) bool {
	for _, r := range s {
		if r >= '0' && r <= '9' {
			return true
		}
	}
	return false
}

// lastNumber extracts the final numeric token on the line, treating dots and
// commas as thousand separators (Rupiah integers).
func lastNumber(line string) int64 {
	matches := numberToken.FindAllString(line, -1)
	if len(matches) == 0 {
		return 0
	}
	return parseAmount(matches[len(matches)-1])
}

func parseAmount(token string) int64 {
	clean := strings.Map(func(r rune) rune {
		if r == '.' || r == ',' || r == ' ' {
			return -1
		}
		return r
	}, token)
	var v int64
	for _, r := range clean {
		if r < '0' || r > '9' {
			continue
		}
		v = v*10 + int64(r-'0')
		if v > math.MaxInt64/10 {
			return v
		}
	}
	return v
}

func parseItemLine(line string) (ParsedItem, bool) {
	numbers := numberToken.FindAllString(line, -1)
	if len(numbers) == 0 {
		return ParsedItem{}, false
	}

	name := numberToken.ReplaceAllString(line, " ")
	name = wsRe.ReplaceAllString(strings.TrimSpace(strings.Trim(name, "-Â·â€¢")), " ")
	confidence := 0.9

	quantity := 1
	unitPrice := int64(0)
	totalPrice := lastNumber(line)

	// Pattern: "Nasi Goreng 2x25.000 ..." or "... 2 @ 12.500"
	if m := qtyXUnit.FindStringSubmatch(line); m != nil {
		quantity = clampQty(parseDigits(m[1]))
		unitPrice = parseAmount(m[2])
	} else if len(numbers) >= 2 {
		// Pattern: "Ayam Goreng 2 15.000" -> second-to-last may be qty
		if q := parseDigits(numbers[len(numbers)-2]); q >= 1 && q <= 99 {
			quantity = q
			unitPrice = parseAmount(numbers[len(numbers)-1])
			totalPrice = unitPrice * int64(quantity)
			confidence = 0.7
		} else {
			unitPrice = parseAmount(numbers[len(numbers)-2])
		}
	} else {
		unitPrice = totalPrice
		confidence = 0.6
	}

	if unitPrice == 0 {
		unitPrice = totalPrice
	}
	if totalPrice == 0 {
		totalPrice = unitPrice * int64(quantity)
	}

	if len(name) < 2 {
		confidence = 0.4
	}

	return ParsedItem{
		Name:       name,
		Quantity:   quantity,
		UnitPrice:  unitPrice,
		TotalPrice: totalPrice,
		Confidence: confidence,
	}, true
}

func clampQty(q int) int {
	if q < 1 {
		return 1
	}
	if q > 999 {
		return 999
	}
	return q
}

func parseDigits(s string) int {
	var v int
	for _, r := range s {
		if r < '0' || r > '9' {
			continue
		}
		v = v*10 + int(r-'0')
	}
	return v
}
