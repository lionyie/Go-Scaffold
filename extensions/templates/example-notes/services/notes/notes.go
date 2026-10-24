// Package notes is a complete example extension: a generated proto service, an
// in-memory implementation, and pages in both apps. Installing it never edits
// hand-written server or router files.
package notes

import (
	"context"
	"sync"

	kerrors "github.com/go-kratos/kratos/v3/errors"
	khttp "github.com/go-kratos/kratos/v3/transport/http"

	notesv1 "github.com/example/scaffold/services/api/notes/v1"
	"github.com/example/scaffold/services/internal/extensions/registry"
	"github.com/example/scaffold/services/internal/pkg/id"
)

func init() {
	registry.Add(func(server *khttp.Server) {
		notesv1.RegisterNotesServiceHTTPServer(server, &service{})
	})
}

type service struct {
	notesv1.UnimplementedNotesServiceServer

	mu    sync.Mutex
	notes []*notesv1.Note
}

func (s *service) ListNotes(_ context.Context, _ *notesv1.ListNotesRequest) (*notesv1.ListNotesResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	items := make([]*notesv1.Note, len(s.notes))
	copy(items, s.notes)
	return &notesv1.ListNotesResponse{Items: items}, nil
}

func (s *service) CreateNote(_ context.Context, req *notesv1.CreateNoteRequest) (*notesv1.CreateNoteResponse, error) {
	note := req.GetNote()
	if note == nil || note.GetTitle() == "" {
		return nil, kerrors.BadRequest("INVALID_ARGUMENT", "title is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	note.Id = id.New()
	s.notes = append(s.notes, note)
	return &notesv1.CreateNoteResponse{Note: note}, nil
}
