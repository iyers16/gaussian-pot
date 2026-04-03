package game

import (
	"math"
	"testing"
)

func TestResolveDebts_NetSumsToZero(t *testing.T) {
	nets := []PlayerNet{
		{UserID: 1, Username: "alice", Net: 50},
		{UserID: 2, Username: "bob", Net: -30},
		{UserID: 3, Username: "carol", Net: -20},
	}
	edges := ResolveDebts(nets)

	// Sum of all edge amounts should equal sum of positive nets.
	totalEdge := 0.0
	for _, e := range edges {
		totalEdge += e.Amount
	}
	totalCredit := 50.0
	if math.Abs(totalEdge-totalCredit) > 0.01 {
		t.Errorf("edge total %.2f != credit total %.2f", totalEdge, totalCredit)
	}
}

func TestResolveDebts_PayerIsDebtor(t *testing.T) {
	nets := []PlayerNet{
		{UserID: 1, Username: "winner", Net: 100},
		{UserID: 2, Username: "loser", Net: -100},
	}
	edges := ResolveDebts(nets)
	if len(edges) != 1 {
		t.Fatalf("expected 1 edge, got %d", len(edges))
	}
	if edges[0].PayerID != 2 {
		t.Errorf("payer should be loser (id=2), got %d", edges[0].PayerID)
	}
	if edges[0].PayeeID != 1 {
		t.Errorf("payee should be winner (id=1), got %d", edges[0].PayeeID)
	}
	if math.Abs(edges[0].Amount-100) > 0.01 {
		t.Errorf("expected amount 100, got %.2f", edges[0].Amount)
	}
}

func TestResolveDebts_AllZero(t *testing.T) {
	nets := []PlayerNet{
		{UserID: 1, Username: "a", Net: 0},
		{UserID: 2, Username: "b", Net: 0},
	}
	edges := ResolveDebts(nets)
	if len(edges) != 0 {
		t.Errorf("expected no edges for zero nets, got %d", len(edges))
	}
}

func TestResolveDebts_MultipleDebtors(t *testing.T) {
	// One big winner, three small losers.
	nets := []PlayerNet{
		{UserID: 1, Username: "big_win", Net: 150},
		{UserID: 2, Username: "lose1", Net: -50},
		{UserID: 3, Username: "lose2", Net: -60},
		{UserID: 4, Username: "lose3", Net: -40},
	}
	edges := ResolveDebts(nets)

	// All loser amounts should total 150.
	total := 0.0
	for _, e := range edges {
		total += e.Amount
	}
	if math.Abs(total-150) > 0.01 {
		t.Errorf("total payments %.2f != 150", total)
	}
}

func TestResolveDebts_Empty(t *testing.T) {
	edges := ResolveDebts(nil)
	if len(edges) != 0 {
		t.Errorf("expected no edges for nil input, got %d", len(edges))
	}
}

func TestResolveDebts_Minimises_Edges(t *testing.T) {
	// Classic case: greedy produces the minimum number of transfers.
	nets := []PlayerNet{
		{UserID: 1, Username: "a", Net: 50},
		{UserID: 2, Username: "b", Net: 50},
		{UserID: 3, Username: "c", Net: -50},
		{UserID: 4, Username: "d", Net: -50},
	}
	edges := ResolveDebts(nets)
	// Each debtor pays exactly one creditor → 2 edges.
	if len(edges) > 2 {
		t.Errorf("expected at most 2 edges, got %d", len(edges))
	}
}
