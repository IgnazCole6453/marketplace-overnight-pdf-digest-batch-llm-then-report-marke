package digest

import "strings"

type SellerAsset struct {
	Title string
	Stock int
}

type BuyerUpdate struct {
	OrderID string
	Message string
}

type OrderHandoff struct {
	OrderID string
	Status  string
}

type Backlog struct {
	SellerAssets []SellerAsset
	BuyerUpdates []BuyerUpdate
	Handoffs     []OrderHandoff
}

// NeedsBuyerAttention keeps the digest focused on handoffs that need a reply.
func NeedsBuyerAttention(update BuyerUpdate, handoffs []OrderHandoff) bool {
	if strings.TrimSpace(update.Message) == "" {
		return false
	}
	for _, handoff := range handoffs {
		if handoff.OrderID == update.OrderID && handoff.Status != "complete" {
			return true
		}
	}
	return false
}

func Prompt(backlog Backlog) string {
	var lines []string
	lines = append(lines, "Summarise this marketplace backlog for the morning operations digest.")
	for _, asset := range backlog.SellerAssets {
		lines = append(lines, "seller asset: "+asset.Title)
	}
	for _, update := range backlog.BuyerUpdates {
		if NeedsBuyerAttention(update, backlog.Handoffs) {
			lines = append(lines, "buyer update for "+update.OrderID+": "+update.Message)
		}
	}
	for _, handoff := range backlog.Handoffs {
		lines = append(lines, "order handoff: "+handoff.OrderID+" is "+handoff.Status)
	}
	return strings.Join(lines, "\n")
}
