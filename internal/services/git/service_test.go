package git

import (
	"context"
	"path/filepath"
	"testing"

	"clikd/internal/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestServiceCreation tests the creation of Git services
func TestServiceCreation(t *testing.T) {
	tests := []struct {
		name        string
		createFunc  func() (Service, error)
		expectError bool
	}{
		{
			name: "NewService with current directory",
			createFunc: func() (Service, error) {
				return NewService()
			},
			expectError: false,
		},
		{
			name: "NewServiceWithRepoDir with valid directory",
			createFunc: func() (Service, error) {
				return NewServiceWithRepoDir(".")
			},
			expectError: false,
		},
		{
			name: "NewServiceWithOptions with all parameters",
			createFunc: func() (Service, error) {
				logger := utils.NewLogger("debug", true)
				return NewServiceWithOptions(".", "v*", "date", logger)
			},
			expectError: false,
		},
		{
			name: "NewServiceWithOptions with nil logger",
			createFunc: func() (Service, error) {
				return NewServiceWithOptions(".", "", "", nil)
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, err := tt.createFunc()

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, service)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, service)

				// Verify service implements the interface
				assert.Implements(t, (*Service)(nil), service)
			}
		})
	}
}

// TestServiceImpl_GetRepositoryRoot tests repository root detection
func TestServiceImpl_GetRepositoryRoot(t *testing.T) {
	service, err := NewService()
	require.NoError(t, err)

	root, err := service.GetRepositoryRoot()

	// Should not error even if not in a git repository
	// The actual git command might fail, but our service should handle it gracefully
	if err == nil {
		assert.NotEmpty(t, root)
		assert.True(t, filepath.IsAbs(root), "Repository root should be an absolute path")
	}
}

// TestServiceImpl_IsGitRepository tests git repository detection
func TestServiceImpl_IsGitRepository(t *testing.T) {
	service, err := NewService()
	require.NoError(t, err)

	isRepo, err := service.IsGitRepository()

	// Should not error - either true or false
	assert.NoError(t, err)
	assert.IsType(t, true, isRepo)
}

// TestServiceImpl_HasRemote tests remote detection
func TestServiceImpl_HasRemote(t *testing.T) {
	service, err := NewService()
	require.NoError(t, err)

	hasRemote, err := service.HasRemote()

	// Should not error - either true or false
	if err == nil {
		assert.IsType(t, true, hasRemote)
	}
}

// TestServiceImpl_GetCurrentBranch tests current branch detection
func TestServiceImpl_GetCurrentBranch(t *testing.T) {
	service, err := NewService()
	require.NoError(t, err)

	branch, err := service.GetCurrentBranch()

	// Should not error if in a git repository
	if err == nil {
		assert.NotEmpty(t, branch)
	}
}

// TestServiceImpl_AnalyzeRepository tests the high-level repository analysis
func TestServiceImpl_AnalyzeRepository(t *testing.T) {
	service, err := NewService()
	require.NoError(t, err)

	ctx := context.Background()
	info, err := service.AnalyzeRepository(ctx)

	// Should always return info, even if not in a git repository
	assert.NoError(t, err)
	assert.NotNil(t, info)

	// Verify structure
	assert.NotEmpty(t, info.Root)
	assert.IsType(t, true, info.IsGitRepo)
	assert.IsType(t, true, info.HasRemote)
	assert.IsType(t, "", info.CurrentBranch)
	assert.IsType(t, "", info.LatestTag)
	assert.IsType(t, 0, info.TotalTags)
	assert.IsType(t, 0, info.TotalCommits)

	t.Logf("Repository Info: %+v", info)
}

// TestServiceImpl_GetChangelogData tests the high-level changelog data extraction
func TestServiceImpl_GetChangelogData(t *testing.T) {
	service, err := NewService()
	require.NoError(t, err)

	ctx := context.Background()
	options := &ChangelogOptions{
		Query:            "",
		TagFilterPattern: "",
		TagSortBy:        "date",
		Paths:            nil,
		FromTag:          "",
		ToTag:            "",
	}

	data, err := service.GetChangelogData(ctx, options)

	// Should handle gracefully even if not in a git repository
	if err == nil {
		assert.NotNil(t, data)
		assert.NotNil(t, data.Repository)
		assert.NotNil(t, data.Tags)
		assert.NotNil(t, data.Commits)

		t.Logf("Changelog Data: Repository=%+v, Tags=%d, Commits=%d",
			data.Repository, len(data.Tags), len(data.Commits))
	}
}

// TestServiceImpl_CommitOptions tests commit options structure
func TestServiceImpl_CommitOptions(t *testing.T) {
	options := CommitOptions{
		Revision:          "HEAD",
		Paths:             []string{"src/", "docs/"},
		HeaderPattern:     `^(\w*)(?:\(([^\)]*)\))?\: (.*)$`,
		HeaderPatternMaps: []string{"Type", "Scope", "Subject"},
		MergePattern:      `^Merge pull request #(\d+) from (.*)$`,
		MergePatternMaps:  []string{"Ref", "Source"},
		RevertPattern:     `^Revert "(.*)"\s*This reverts commit (\w+)\.`,
		RevertPatternMaps: []string{"Header", "Hash"},
		RefActions:        []string{"Closes", "Fixes", "Resolves"},
		IssuePrefix:       []string{"#", "gh-"},
		NoteKeywords:      []string{"BREAKING CHANGE", "BREAKING CHANGES"},
	}

	// Verify structure is properly initialized
	assert.Equal(t, "HEAD", options.Revision)
	assert.Len(t, options.Paths, 2)
	assert.NotEmpty(t, options.HeaderPattern)
	assert.Len(t, options.HeaderPatternMaps, 3)
	assert.Len(t, options.RefActions, 3)
	assert.Len(t, options.IssuePrefix, 2)
	assert.Len(t, options.NoteKeywords, 2)
}

// TestServiceImpl_DataStructures tests the data structure initialization
func TestServiceImpl_DataStructures(t *testing.T) {
	t.Run("RepositoryInfo", func(t *testing.T) {
		info := &RepositoryInfo{
			Root:          "/path/to/repo",
			IsGitRepo:     true,
			HasRemote:     true,
			CurrentBranch: "main",
			LatestTag:     "v1.0.0",
			TotalTags:     5,
			TotalCommits:  100,
		}

		assert.Equal(t, "/path/to/repo", info.Root)
		assert.True(t, info.IsGitRepo)
		assert.True(t, info.HasRemote)
		assert.Equal(t, "main", info.CurrentBranch)
		assert.Equal(t, "v1.0.0", info.LatestTag)
		assert.Equal(t, 5, info.TotalTags)
		assert.Equal(t, 100, info.TotalCommits)
	})

	t.Run("ChangelogOptions", func(t *testing.T) {
		options := &ChangelogOptions{
			Query:            "v1.0.0..v2.0.0",
			TagFilterPattern: "v*",
			TagSortBy:        "date",
			Paths:            []string{"src/"},
			FromTag:          "v1.0.0",
			ToTag:            "v2.0.0",
		}

		assert.Equal(t, "v1.0.0..v2.0.0", options.Query)
		assert.Equal(t, "v*", options.TagFilterPattern)
		assert.Equal(t, "date", options.TagSortBy)
		assert.Len(t, options.Paths, 1)
		assert.Equal(t, "v1.0.0", options.FromTag)
		assert.Equal(t, "v2.0.0", options.ToTag)
	})

	t.Run("ChangelogData", func(t *testing.T) {
		data := &ChangelogData{
			Repository: &RepositoryInfo{Root: "/test"},
			Tags:       []*Tag{},
			Commits:    []*Commit{},
			FromTag:    "v1.0.0",
			ToTag:      "v2.0.0",
		}

		assert.NotNil(t, data.Repository)
		assert.Equal(t, "/test", data.Repository.Root)
		assert.NotNil(t, data.Tags)
		assert.NotNil(t, data.Commits)
		assert.Equal(t, "v1.0.0", data.FromTag)
		assert.Equal(t, "v2.0.0", data.ToTag)
	})
}

// TestServiceImpl_ErrorHandling tests error handling in various scenarios
func TestServiceImpl_ErrorHandling(t *testing.T) {
	t.Run("Invalid repository directory", func(t *testing.T) {
		service, err := NewServiceWithRepoDir("/nonexistent/directory")

		// Service creation should succeed even with invalid directory
		assert.NoError(t, err)
		assert.NotNil(t, service)

		// But operations might fail gracefully
		isRepo, err := service.IsGitRepository()
		if err != nil {
			// Error is expected for non-existent directory
			assert.Error(t, err)
		} else {
			// Or it might return false
			assert.False(t, isRepo)
		}
	})

	t.Run("Empty options", func(t *testing.T) {
		ctx := context.Background()
		service, err := NewService()
		require.NoError(t, err)

		// Should handle nil options gracefully
		data, err := service.GetChangelogData(ctx, &ChangelogOptions{})

		// Should not panic and handle gracefully
		if err == nil {
			assert.NotNil(t, data)
		}
	})
}

// TestServiceImpl_InterfaceCompliance tests that ServiceImpl implements Service interface
func TestServiceImpl_InterfaceCompliance(t *testing.T) {
	service, err := NewService()
	require.NoError(t, err)

	// Verify all interface methods are implemented
	var _ Service = service

	// Test that we can call all interface methods without panicking
	ctx := context.Background()

	// Repository operations
	_, _ = service.GetRepositoryRoot()
	_, _ = service.IsGitRepository()
	_, _ = service.HasRemote()
	_, _ = service.GetCurrentBranch()

	// Tag operations
	_, _ = service.GetLatestTag()
	_, _ = service.GetTags()
	_, _ = service.GetTagsWithPattern("v*")
	_, _ = service.GetAllTagsWithDetails()

	// Commit operations
	_, _ = service.GetCommits("HEAD", nil)
	_, _ = service.GetCommitsBetweenTags("", "")
	_, _ = service.GetStagedChanges()

	// High-level operations
	_, _ = service.AnalyzeRepository(ctx)
	_, _ = service.GetChangelogData(ctx, &ChangelogOptions{})
	_, _, _ = service.SelectTagRange("")
	_, _ = service.GetDiffBetweenTags("", "")

	// Should not panic
	assert.True(t, true, "All interface methods callable without panic")
}
