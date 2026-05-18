package reader

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"recently_played/internal/cassandra"
	"recently_played/internal/config"
)

type fakeDB struct {
	results []cassandra.Entry
	err     error

	readCalls int
	gotCtx    context.Context
	gotUserID string
	gotLimit  int
}

func (f *fakeDB) Read(ctx context.Context, userID string, limit int) ([]cassandra.Entry, error) {
	f.readCalls++
	f.gotCtx = ctx
	f.gotUserID = userID
	f.gotLimit = limit

	return f.results, f.err
}

func (f *fakeDB) Write(context.Context, cassandra.Entry) error { return nil }
func (f *fakeDB) Close()                                       {}

func TestService_GetLastPlayedTracks(t *testing.T) {
	tests := []struct {
		name    string
		cfg     config.Config
		userID  string
		results []cassandra.Entry
		dbErr   error

		want    []Entry
		wantErr bool
	}{
		{
			name: "returns db error",
			cfg: config.Config{
				RecentTracksCount: 2,
				DatabaseReadCount: 10,
			},
			userID:  "user-1",
			dbErr:   errors.New("db failed"),
			wantErr: true,
		},
		{
			name: "deduplicates by track id and respects limit",
			cfg: config.Config{
				RecentTracksCount: 3,
				DatabaseReadCount: 10,
			},
			userID: "user-42",
			results: []cassandra.Entry{
				{UserId: "ignored", PlayedAt: 500, TrackId: "track-a"},
				{UserId: "ignored", PlayedAt: 400, TrackId: "track-b"},
				{UserId: "ignored", PlayedAt: 300, TrackId: "track-a"},
				{UserId: "ignored", PlayedAt: 200, TrackId: "track-c"},
				{UserId: "ignored", PlayedAt: 100, TrackId: "track-d"},
			},
			want: []Entry{
				{UserId: "user-42", PlayedAt: 500, TrackId: "track-a"},
				{UserId: "user-42", PlayedAt: 400, TrackId: "track-b"},
				{UserId: "user-42", PlayedAt: 200, TrackId: "track-c"},
			},
		},
		{
			name: "returns all unique entries when fewer than limit",
			cfg: config.Config{
				RecentTracksCount: 5,
				DatabaseReadCount: 20,
			},
			userID: "user-99",
			results: []cassandra.Entry{
				{UserId: "other", PlayedAt: 900, TrackId: "track-x"},
				{UserId: "other", PlayedAt: 800, TrackId: "track-y"},
				{UserId: "other", PlayedAt: 700, TrackId: "track-x"},
			},
			want: []Entry{
				{UserId: "user-99", PlayedAt: 900, TrackId: "track-x"},
				{UserId: "user-99", PlayedAt: 800, TrackId: "track-y"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeDB{results: tt.results, err: tt.dbErr}
			s := New(tt.cfg, fake)

			ctx := context.WithValue(context.Background(), "request-id", "abc-123")
			got, err := s.GetLastPlayedTracks(ctx, tt.userID)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("unexpected result\nwant: %#v\n got: %#v", tt.want, got)
			}

			if fake.readCalls != 1 {
				t.Fatalf("expected Read to be called once, got %d", fake.readCalls)
			}
			if fake.gotCtx != ctx {
				t.Fatalf("expected context to be passed to Read")
			}
			if fake.gotUserID != tt.userID {
				t.Fatalf("expected userID %q, got %q", tt.userID, fake.gotUserID)
			}
			if fake.gotLimit != tt.cfg.DatabaseReadCount {
				t.Fatalf("expected read limit %d, got %d", tt.cfg.DatabaseReadCount, fake.gotLimit)
			}
		})
	}
}
