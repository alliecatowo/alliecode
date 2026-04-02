package tools

import "testing"

func TestApplyNotebookCellActionUpdateAddDelete(t *testing.T) {
	nb := notebookFile{Cells: []notebookCell{{CellType: "code", Source: []string{"a\n"}}}}

	err := applyNotebookCellAction(&nb, notebookEditInput{CellIndex: 0, NewSource: "print(1)"}, "update")
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if len(nb.Cells[0].Source) == 0 {
		t.Fatalf("expected updated source")
	}

	err = applyNotebookCellAction(&nb, notebookEditInput{CellIndex: 1, NewSource: "# note", CellType: "markdown"}, "add")
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}
	if len(nb.Cells) != 2 || nb.Cells[1].CellType != "markdown" {
		t.Fatalf("expected markdown cell appended, got %#v", nb.Cells)
	}

	err = applyNotebookCellAction(&nb, notebookEditInput{CellIndex: 0}, "delete")
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if len(nb.Cells) != 1 {
		t.Fatalf("expected one cell after delete, got %d", len(nb.Cells))
	}
}

func TestApplyNotebookCellActionValidation(t *testing.T) {
	nb := notebookFile{Cells: []notebookCell{{CellType: "code"}}}

	if err := applyNotebookCellAction(&nb, notebookEditInput{CellIndex: 0}, "update"); err == nil {
		t.Fatalf("expected error for missing new_source on update")
	}
	if err := applyNotebookCellAction(&nb, notebookEditInput{CellIndex: 5, NewSource: "x"}, "add"); err == nil {
		t.Fatalf("expected out of range error for add")
	}
	if err := applyNotebookCellAction(&nb, notebookEditInput{CellIndex: 0}, "unknown"); err == nil {
		t.Fatalf("expected error for invalid action")
	}
}
