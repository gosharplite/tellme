package llm

import (
	"math"
	"testing"
)

func TestComputeCost(t *testing.T) {
	p := Pricing{Hit: 0.0028, Miss: 0.14, Comp: 0.28}
	got := ComputeCost(p, 4, 6, 3, 2) // miss 4, hit 6, completion 3, thinking 2
	want := (4*0.14 + 6*0.0028 + (3+2)*0.28) / 1_000_000
	if math.Abs(got-want) > 1e-15 {
		t.Errorf("ComputeCost = %v, want %v", got, want)
	}
	if ComputeCost(p, 4, 6, 3, 5) <= ComputeCost(p, 4, 6, 3, 2) {
		t.Errorf("thinking tokens must be billed at the completion rate")
	}
}

func TestComputeCostUnpricedIsZero(t *testing.T) {
	if got := ComputeCost(Pricing{}, 4, 6, 3, 2); got != 0 {
		t.Errorf("an un-priced model must cost 0, got %v", got)
	}
}

func TestHitRate(t *testing.T) {
	if got := HitRate(6, 4); math.Abs(got-60.0) > 1e-9 {
		t.Errorf("HitRate(6,4) = %v, want 60", got)
	}
	if got := HitRate(0, 0); got != 0 {
		t.Errorf("HitRate(0,0) = %v, want 0", got)
	}
}
