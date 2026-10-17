package digest

import "testing"

func TestNeedsBuyerAttention(t *testing.T) {
	handoffs := []OrderHandoff{{OrderID: "o-17", Status: "awaiting_seller"}, {OrderID: "o-18", Status: "complete"}}
	tests := []struct {
		name   string
		update BuyerUpdate
		want   bool
	}{
		{name: "open handoff with buyer message", update: BuyerUpdate{OrderID: "o-17", Message: "Please confirm dispatch."}, want: true},
		{name: "completed handoff", update: BuyerUpdate{OrderID: "o-18", Message: "Thanks."}, want: false},
		{name: "empty update", update: BuyerUpdate{OrderID: "o-17"}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NeedsBuyerAttention(tt.update, handoffs); got != tt.want {
				t.Fatalf("NeedsBuyerAttention() = %v, want %v", got, tt.want)
			}
		})
	}
}
