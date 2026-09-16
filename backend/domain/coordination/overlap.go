package coordination

import (
	"sort"
	"strings"
)

const (
	TitleSelfOverlap = "Parallel Slot Overlap"
	TitlePeerOverlap = "Peer Scope Overlap"
)

// Overlap is a derived collision between claims. Not stored; Pulse Conflict is a DTO of this.
type Overlap struct {
	Severity    string
	Title       string
	Description string
	Aliases     []string
	Service     string
}

type slotKey struct {
	person, taskID string
}

type slotClaims struct {
	person, taskID, title string
	claims                []Claim
}

func groupSlots(claims []Claim) []slotClaims {
	order := make([]slotKey, 0)
	by := make(map[slotKey]*slotClaims)
	for _, c := range claims {
		if c.Person == "" || c.TaskID == "" {
			continue
		}
		k := slotKey{c.Person, c.TaskID}
		s, ok := by[k]
		if !ok {
			s = &slotClaims{person: c.Person, taskID: c.TaskID, title: c.Title}
			by[k] = s
			order = append(order, k)
		}
		s.claims = append(s.claims, c)
	}
	out := make([]slotClaims, 0, len(order))
	for _, k := range order {
		out = append(out, *by[k])
	}
	return out
}

func firstRelated(a, b []Claim, allow func(Resource) bool) (Resource, bool) {
	for _, left := range a {
		if !allow(left.Resource) {
			continue
		}
		for _, right := range b {
			if !allow(right.Resource) {
				continue
			}
			if left.Resource.Relates(right.Resource) {
				return right.Resource, true
			}
		}
	}
	return Resource{}, false
}

func selfAllowed(r Resource) bool {
	return r.Kind == KindProduct || r.Kind == KindWorkspace
}

func peerAllowed(r Resource) bool {
	return r.Kind == KindProduct
}

// SelfOverlaps is exclusive policy: one person, two tasks, same product or workspace zone.
func SelfOverlaps(claims []Claim) []Overlap {
	slots := groupSlots(claims)
	out := make([]Overlap, 0)
	for i := 0; i < len(slots); i++ {
		for j := i + 1; j < len(slots); j++ {
			if slots[i].person != slots[j].person {
				continue
			}
			res, ok := firstRelated(slots[i].claims, slots[j].claims, selfAllowed)
			if !ok {
				continue
			}
			label := res.Label
			out = append(out, Overlap{
				Severity:    "critical",
				Title:       TitleSelfOverlap,
				Description: slots[i].person + " has two in-progress tasks claiming " + label,
				Aliases:     []string{slots[i].person},
				Service:     label,
			})
		}
	}
	return out
}

// PeerOverlaps is advisory policy: two people, same product zone.
func PeerOverlaps(claims []Claim) []Overlap {
	out := make([]Overlap, 0)
	seen := make(map[string]struct{})
	for i := 0; i < len(claims); i++ {
		a := claims[i]
		if !peerAllowed(a.Resource) {
			continue
		}
		for j := i + 1; j < len(claims); j++ {
			b := claims[j]
			if a.Person == b.Person || !peerAllowed(b.Resource) || !a.Resource.Relates(b.Resource) {
				continue
			}
			fp := peerPairKey(a, b)
			if _, ok := seen[fp]; ok {
				continue
			}
			seen[fp] = struct{}{}
			aliases := []string{a.Person, b.Person}
			sort.Strings(aliases)
			out = append(out, Overlap{
				Severity:    "warning",
				Title:       TitlePeerOverlap,
				Description: a.Person + " «" + a.Title + "» and " + b.Person + " «" + b.Title + "» both claim " + a.Resource.Label,
				Aliases:     aliases,
				Service:     a.Resource.Label,
			})
		}
	}
	return out
}

func peerPairKey(a, b Claim) string {
	left, right := a, b
	if a.Person > b.Person || (a.Person == b.Person && a.TaskID > b.TaskID) {
		left, right = b, a
	}
	return strings.ToLower(left.Person + ":" + left.TaskID + "|" + right.Person + ":" + right.TaskID + "|" + left.Resource.Key)
}
