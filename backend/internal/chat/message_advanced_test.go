package chat

import (
	"testing"
	"time"
)

func TestBranchManagerCreateBranch(t *testing.T) {
	bm := NewBranchManager()

	branch, err := bm.CreateBranch("session-1", "Alternative Path", "msg-1")
	if err != nil {
		t.Errorf("CreateBranch failed: %v", err)
	}

	if branch.Name != "Alternative Path" {
		t.Errorf("Expected Alternative Path")
	}

	if branch.ParentMessageID != "msg-1" {
		t.Errorf("Expected parent msg-1")
	}
}

func TestBranchManagerGetBranch(t *testing.T) {
	bm := NewBranchManager()

	branch, _ := bm.CreateBranch("session-1", "Branch", "msg-1")

	retrieved, err := bm.GetBranch(branch.ID)
	if err != nil {
		t.Errorf("GetBranch failed: %v", err)
	}

	if retrieved.Name != "Branch" {
		t.Errorf("Expected Branch")
	}
}

func TestBranchManagerAddMessageToBranch(t *testing.T) {
	bm := NewBranchManager()

	branch, _ := bm.CreateBranch("session-1", "Branch", "msg-1")

	err := bm.AddMessageToBranch(branch.ID, "msg-2")
	if err != nil {
		t.Errorf("AddMessageToBranch failed: %v", err)
	}

	retrieved, _ := bm.GetBranch(branch.ID)
	if len(retrieved.MessageIDs) != 1 {
		t.Errorf("Expected 1 message in branch")
	}
}

func TestBranchManagerGetBranchesForMessage(t *testing.T) {
	bm := NewBranchManager()

	bm.CreateBranch("session-1", "Branch1", "msg-1")
	bm.CreateBranch("session-1", "Branch2", "msg-1")

	branches := bm.GetBranchesForMessage("msg-1")

	if len(branches) != 2 {
		t.Errorf("Expected 2 branches for message")
	}
}

func TestEditManagerRecordEdit(t *testing.T) {
	em := NewEditManager()

	edit, err := em.RecordEdit("msg-1", "Original", "Modified", "user-1", "Typo fix")
	if err != nil {
		t.Errorf("RecordEdit failed: %v", err)
	}

	if edit.NewContent != "Modified" {
		t.Errorf("Expected Modified content")
	}
}

func TestEditManagerGetEditHistory(t *testing.T) {
	em := NewEditManager()

	em.RecordEdit("msg-1", "Original", "Version2", "user-1", "Reason1")
	em.RecordEdit("msg-1", "Version2", "Version3", "user-2", "Reason2")

	history := em.GetEditHistory("msg-1")

	if len(history) != 2 {
		t.Errorf("Expected 2 edits in history")
	}
}

func TestEditManagerGetLatestEdit(t *testing.T) {
	em := NewEditManager()

	em.RecordEdit("msg-1", "Original", "Version2", "user-1", "Reason1")
	em.RecordEdit("msg-1", "Version2", "Version3", "user-2", "Reason2")

	latest, err := em.GetLatestEdit("msg-1")
	if err != nil {
		t.Errorf("GetLatestEdit failed: %v", err)
	}

	if latest.NewContent != "Version3" {
		t.Errorf("Expected Version3")
	}
}

func TestReferenceManagerCreateReference(t *testing.T) {
	rm := NewReferenceManager()

	ref, err := rm.CreateReference("msg-1", "msg-2", "See also msg-1")
	if err != nil {
		t.Errorf("CreateReference failed: %v", err)
	}

	if ref.ReferencedMessageID != "msg-1" {
		t.Errorf("Expected msg-1 as referenced")
	}
}

func TestReferenceManagerGetReferences(t *testing.T) {
	rm := NewReferenceManager()

	rm.CreateReference("msg-1", "msg-2", "Context1")
	rm.CreateReference("msg-3", "msg-2", "Context2")

	refs := rm.GetReferences("msg-2")

	if len(refs) != 2 {
		t.Errorf("Expected 2 references")
	}
}

func TestReferenceManagerGetReferencedBy(t *testing.T) {
	rm := NewReferenceManager()

	rm.CreateReference("msg-1", "msg-2", "Context")
	rm.CreateReference("msg-1", "msg-3", "Context")

	refs := rm.GetReferencedBy("msg-1")

	if len(refs) > 0 {
		t.Logf("Found %d references", len(refs))
	}
}

func TestExportManagerExportJSON(t *testing.T) {
	em := NewExportManager()

	messages := []*Message{
		{
			ID:        "msg-1",
			Role:      "user",
			Content:   "Hello",
			Tokens:    5,
			Timestamp: time.Now(),
		},
	}

	export, err := em.ExportMessages("session-1", messages, ExportFormatJSON)
	if err != nil {
		t.Errorf("ExportMessages failed: %v", err)
	}

	if export.Format != ExportFormatJSON {
		t.Errorf("Expected JSON format")
	}

	if export.MessageCount != 1 {
		t.Errorf("Expected 1 message in export")
	}
}

func TestExportManagerExportMarkdown(t *testing.T) {
	em := NewExportManager()

	messages := []*Message{
		{
			ID:        "msg-1",
			Role:      "user",
			Content:   "Hello",
			Tokens:    5,
			Timestamp: time.Now(),
		},
	}

	export, err := em.ExportMessages("session-1", messages, ExportFormatMarkdown)
	if err != nil {
		t.Errorf("ExportMessages failed: %v", err)
	}

	if export.Format != ExportFormatMarkdown {
		t.Errorf("Expected Markdown format")
	}
}

func TestExportManagerExportHTML(t *testing.T) {
	em := NewExportManager()

	messages := []*Message{
		{
			ID:        "msg-1",
			Role:      "user",
			Content:   "Hello",
			Tokens:    5,
			Timestamp: time.Now(),
		},
	}

	export, err := em.ExportMessages("session-1", messages, ExportFormatHTML)
	if err != nil {
		t.Errorf("ExportMessages failed: %v", err)
	}

	if export.Format != ExportFormatHTML {
		t.Errorf("Expected HTML format")
	}
}

func TestExportManagerExportCSV(t *testing.T) {
	em := NewExportManager()

	messages := []*Message{
		{
			ID:        "msg-1",
			Role:      "user",
			Content:   "Hello",
			Tokens:    5,
			Timestamp: time.Now(),
		},
	}

	export, err := em.ExportMessages("session-1", messages, ExportFormatCSV)
	if err != nil {
		t.Errorf("ExportMessages failed: %v", err)
	}

	if export.Format != ExportFormatCSV {
		t.Errorf("Expected CSV format")
	}
}

func TestExportManagerGetExport(t *testing.T) {
	em := NewExportManager()

	messages := []*Message{
		{
			ID:        "msg-1",
			Role:      "user",
			Content:   "Hello",
			Tokens:    5,
			Timestamp: time.Now(),
		},
	}

	export, _ := em.ExportMessages("session-1", messages, ExportFormatJSON)

	retrieved, err := em.GetExport(export.ID)
	if err != nil {
		t.Errorf("GetExport failed: %v", err)
	}

	if retrieved.MessageCount != 1 {
		t.Errorf("Expected 1 message")
	}
}

func TestAdvancedMessageManagerStatistics(t *testing.T) {
	amm := NewAdvancedMessageManager()

	amm.branchManager.CreateBranch("session-1", "Branch", "msg-1")
	amm.editManager.RecordEdit("msg-1", "Original", "Modified", "user-1", "Reason")
	amm.referenceManager.CreateReference("msg-1", "msg-2", "Context")

	stats := amm.GetStatistics()

	if stats == nil {
		t.Errorf("Expected statistics")
	}
}

func BenchmarkCreateBranch(b *testing.B) {
	bm := NewBranchManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bm.CreateBranch("session-1", "Branch", "msg-"+string(rune(i)))
	}
}

func BenchmarkRecordEdit(b *testing.B) {
	em := NewEditManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		em.RecordEdit("msg-1", "Original", "Modified", "user-1", "Reason")
	}
}

func BenchmarkCreateReference(b *testing.B) {
	rm := NewReferenceManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rm.CreateReference("msg-1", "msg-"+string(rune(i)), "Context")
	}
}

func BenchmarkExportJSON(b *testing.B) {
	em := NewExportManager()

	messages := []*Message{
		{
			ID:        "msg-1",
			Role:      "user",
			Content:   "Hello",
			Tokens:    5,
			Timestamp: time.Now(),
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = em.ExportMessages("session-1", messages, ExportFormatJSON)
	}
}

