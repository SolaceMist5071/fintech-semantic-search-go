package main

import "strings"

type PaymentEvent struct {
	ID     string `json:"id"`
	Text   string `json:"text"`
	Amount int64  `json:"amount_cents"`
	Risk   string `json:"risk"`
}

func actionFor(e PaymentEvent) string {
	if strings.EqualFold(e.Risk, "high") || e.Amount >= 1000000 {
		return "review"
	}
	return "allow"
}
