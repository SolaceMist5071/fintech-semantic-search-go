package main

import "testing"

func TestActionForPaymentRisk(t *testing.T) {
	cases := []struct {
		name  string
		event PaymentEvent
		want  string
	}{{"high risk", PaymentEvent{Risk: "high", Amount: 100}, "review"}, {"large amount", PaymentEvent{Risk: "low", Amount: 1000000}, "review"}, {"routine", PaymentEvent{Risk: "low", Amount: 2500}, "allow"}}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := actionFor(tc.event); got != tc.want {
				t.Fatalf("actionFor()=%q,want %q", got, tc.want)
			}
		})
	}
}
