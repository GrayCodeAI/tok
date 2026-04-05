package favorites_test

import (
	"testing"

	"github.com/GrayCodeAI/tokman/internal/favorites"
)

func TestNewFavoritesManager(t *testing.T) {
	fm := favorites.NewFavoritesManager()
	if fm == nil {
		t.Fatal("NewFavoritesManager returned nil")
	}
}

func TestAddFavorite(t *testing.T) {
	fm := favorites.NewFavoritesManager()
	fav := &favorites.FavoriteCommand{
		ID:      "1",
		Name:    "git-status",
		Command: "git status",
		Tags:    []string{"git"},
	}
	if err := fm.AddFavorite(fav); err != nil {
		t.Fatalf("AddFavorite error = %v", err)
	}

	list := fm.ListFavorites()
	if len(list) != 1 {
		t.Errorf("ListFavorites() = %d items, want 1", len(list))
	}
}

func TestGetFavorite(t *testing.T) {
	fm := favorites.NewFavoritesManager()
	fm.AddFavorite(&favorites.FavoriteCommand{
		ID:      "1",
		Name:    "git-log",
		Command: "git log",
	})

	got, err := fm.GetFavorite("1")
	if err != nil {
		t.Fatalf("GetFavorite error = %v", err)
	}
	if got.Command != "git log" {
		t.Errorf("Command = %q, want %q", got.Command, "git log")
	}
}

func TestRemoveFavorite(t *testing.T) {
	fm := favorites.NewFavoritesManager()
	fm.AddFavorite(&favorites.FavoriteCommand{
		ID:      "1",
		Name:    "git-diff",
		Command: "git diff",
	})
	if err := fm.RemoveFavorite("1"); err != nil {
		t.Fatalf("RemoveFavorite error = %v", err)
	}
	if len(fm.ListFavorites()) != 0 {
		t.Error("should have 0 favorites after removal")
	}
}

func TestGetByTag(t *testing.T) {
	fm := favorites.NewFavoritesManager()
	fm.AddFavorite(&favorites.FavoriteCommand{
		ID:      "1",
		Name:    "git-status",
		Command: "git status",
		Tags:    []string{"git"},
	})
	fm.AddFavorite(&favorites.FavoriteCommand{
		ID:      "2",
		Name:    "ls",
		Command: "ls -la",
		Tags:    []string{"list"},
	})

	results := fm.GetByTag("git")
	if len(results) != 1 {
		t.Errorf("GetByTag(git) = %d, want 1", len(results))
	}
}

func TestGetByTag_NoneFound(t *testing.T) {
	fm := favorites.NewFavoritesManager()
	fm.AddFavorite(&favorites.FavoriteCommand{
		ID:      "1",
		Name:    "git-status",
		Command: "git status",
		Tags:    []string{"git"},
	})
	results := fm.GetByTag("nonexistent")
	if len(results) != 0 {
		t.Errorf("GetByTag(nonexistent) = %d, want 0", len(results))
	}
}

func TestRemoveFavorite_NotFound(t *testing.T) {
	fm := favorites.NewFavoritesManager()
	err := fm.RemoveFavorite("nonexistent")
	if err == nil {
		t.Error("expected error when removing nonexistent favorite")
	}
}
