package service

import "testing"

func TestDesiredPermsContainsKeyModes(t *testing.T) {
	perms := DesiredPerms()
	m := map[string]uint32{}
	for _, p := range perms {
		m[p.Path] = uint32(p.Mode)
	}
	// The exact paths are absolute; we just check the required modes appear.
	want := map[uint32]bool{0750: false, 0640: false, 0644: false}
	for _, p := range perms {
		if _, ok := want[uint32(p.Mode)]; ok {
			want[uint32(p.Mode)] = true
		}
	}
	for mode, ok := range want {
		if !ok {
			t.Fatalf("expected mode %04o in DesiredPerms", mode)
		}
	}
}
