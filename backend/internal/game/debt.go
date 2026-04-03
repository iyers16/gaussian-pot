package game

import "math"

// DebtEdge represents a payment obligation from one player to another.
type DebtEdge struct {
	PayerID       int
	PayerUsername string
	PayeeID       int
	PayeeUsername string
	Amount        float64
}

// PlayerNet ties a user to their net position after payouts.
type PlayerNet struct {
	UserID   int
	Username string
	Net      float64 // positive = creditor, negative = debtor
}

// ResolveDebts applies the greedy creditor/debtor algorithm to produce the
// minimal set of P2P payment edges that zero out all net positions.
func ResolveDebts(nets []PlayerNet) []DebtEdge {
	// Separate into creditors (net > 0) and debtors (net < 0).
	creditors := []PlayerNet{}
	debtors := []PlayerNet{}

	for _, p := range nets {
		if p.Net > 0.001 {
			creditors = append(creditors, p)
		} else if p.Net < -0.001 {
			debtors = append(debtors, p)
		}
	}

	var edges []DebtEdge
	ci, di := 0, 0

	for ci < len(creditors) && di < len(debtors) {
		creditor := &creditors[ci]
		debtor := &debtors[di]

		// How much can we settle right now?
		settle := math.Min(creditor.Net, -debtor.Net)
		settle = math.Round(settle*100) / 100 // round to cents

		if settle > 0 {
			edges = append(edges, DebtEdge{
				PayerID:       debtor.UserID,
				PayerUsername: debtor.Username,
				PayeeID:       creditor.UserID,
				PayeeUsername: creditor.Username,
				Amount:        settle,
			})
		}

		creditor.Net -= settle
		debtor.Net += settle

		if math.Abs(creditor.Net) < 0.001 {
			ci++
		}
		if math.Abs(debtor.Net) < 0.001 {
			di++
		}
	}

	return edges
}
