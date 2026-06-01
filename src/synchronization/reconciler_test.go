package synchronization

import (
	"errors"
	"testing"
)

type rItem struct {
	ID    string
	Group string
}

func rEqualID(a, b rItem) bool { return a.ID == b.ID }

func TestReconcile_InsertsRemoteOnlyItems(t *testing.T) {
	ours := []rItem{{ID: "a"}, {ID: "b"}}
	theirs := []rItem{{ID: "a"}, {ID: "c"}, {ID: "d"}}

	var inserted []rItem
	missing, err := reconcile[rItem](
		func() ([]rItem, error) { return ours, nil },
		theirs,
		rEqualID,
		func(i rItem) error { inserted = append(inserted, i); return nil },
		nil,
	)
	if err != nil {
		t.Fatalf("reconcile error: %v", err)
	}
	if len(inserted) != 2 || inserted[0].ID != "c" || inserted[1].ID != "d" {
		t.Errorf("expected inserted=[c,d], got %+v", inserted)
	}
	if len(missing) != 1 || missing[0].ID != "b" {
		t.Errorf("expected missingInRemote=[b], got %+v", missing)
	}
}

func TestReconcile_FullyMatchingSetsNoOp(t *testing.T) {
	items := []rItem{{ID: "a"}, {ID: "b"}}
	var inserted []rItem
	missing, err := reconcile[rItem](
		func() ([]rItem, error) { return items, nil },
		items,
		rEqualID,
		func(i rItem) error { inserted = append(inserted, i); return nil },
		nil,
	)
	if err != nil {
		t.Fatalf("reconcile error: %v", err)
	}
	if len(inserted) != 0 {
		t.Errorf("matching sets should produce no inserts, got %+v", inserted)
	}
	if len(missing) != 0 {
		t.Errorf("matching sets should produce no missingInRemote, got %+v", missing)
	}
}

func TestReconcile_LoaderErrorPropagates(t *testing.T) {
	_, err := reconcile[rItem](
		func() ([]rItem, error) { return nil, errors.New("db gone") },
		nil,
		rEqualID,
		func(rItem) error { return nil },
		nil,
	)
	if err == nil || err.Error() != "db gone" {
		t.Errorf("expected loader error to propagate, got %v", err)
	}
}

func TestReconcile_InsertErrorPropagatesByDefault(t *testing.T) {
	_, err := reconcile[rItem](
		func() ([]rItem, error) { return nil, nil },
		[]rItem{{ID: "a"}},
		rEqualID,
		func(rItem) error { return errors.New("boom") },
		nil,
	)
	if err == nil || err.Error() != "boom" {
		t.Errorf("expected insert error to propagate, got %v", err)
	}
}

func TestReconcile_InsertSkipSwallowsMatchingErrors(t *testing.T) {
	dup := errors.New("duplicate key violation")
	inserts := 0
	missing, err := reconcile[rItem](
		func() ([]rItem, error) { return nil, nil },
		[]rItem{{ID: "a"}, {ID: "b"}},
		rEqualID,
		func(rItem) error { inserts++; return dup },
		func(e error) bool { return e == dup },
	)
	if err != nil {
		t.Fatalf("duplicate-key errors should be swallowed; got %v", err)
	}
	if inserts != 2 {
		t.Errorf("expected 2 insert attempts despite errors, got %d", inserts)
	}
	if len(missing) != 0 {
		t.Errorf("loader returned empty so nothing should be missing, got %+v", missing)
	}
}

func TestReconcile_CustomEqualityForUserGroups(t *testing.T) {
	// Users use ID-or-same-group equality. Two items with different IDs but
	// the same Group should be considered the same item — no insert, no
	// missing-in-remote.
	ours := []rItem{{ID: "user-1", Group: "G"}}
	theirs := []rItem{{ID: "user-2", Group: "G"}}

	equal := func(a, b rItem) bool {
		return a.ID == b.ID || (a.Group != "" && a.Group == b.Group)
	}

	var inserted []rItem
	missing, err := reconcile[rItem](
		func() ([]rItem, error) { return ours, nil },
		theirs,
		equal,
		func(i rItem) error { inserted = append(inserted, i); return nil },
		nil,
	)
	if err != nil {
		t.Fatalf("reconcile error: %v", err)
	}
	if len(inserted) != 0 {
		t.Errorf("same-group items should not insert, got %+v", inserted)
	}
	if len(missing) != 0 {
		t.Errorf("same-group items should not be missing-in-remote, got %+v", missing)
	}
}

func TestIsDuplicateKeyErr_MatchesLegacyWording(t *testing.T) {
	if !isDuplicateKeyErr(errors.New("ERROR: duplicate key value violates ...")) {
		t.Error("expected match on legacy duplicate-key wording")
	}
	if isDuplicateKeyErr(errors.New("some other error")) {
		t.Error("unexpected match on unrelated error")
	}
	if isDuplicateKeyErr(nil) {
		t.Error("nil error should not match")
	}
}
