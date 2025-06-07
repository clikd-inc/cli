package changelog

import (
	"fmt"
	"sort"
	"strings"
)

type commitExtractor struct {
	opts *Options
}

func newCommitExtractor(opts *Options) *commitExtractor {
	return &commitExtractor{
		opts: opts,
	}
}

func (e *commitExtractor) Extract(commits []*Commit) ([]*CommitGroup, []*Commit, []*Commit, []*NoteGroup) {
	fmt.Printf("DEBUG: commitExtractor.Extract called with %d commits\n", len(commits))
	fmt.Printf("DEBUG: CommitGroupBy: %s, CommitSortBy: %s, CommitGroupSortBy: %s\n",
		e.opts.CommitGroupBy, e.opts.CommitSortBy, e.opts.CommitGroupSortBy)
	fmt.Printf("DEBUG: CommitGroupTitleMaps: %v\n", e.opts.CommitGroupTitleMaps)

	commitGroups := []*CommitGroup{}
	noteGroups := []*NoteGroup{}
	mergeCommits := []*Commit{}
	revertCommits := []*Commit{}

	filteredCommits := commitFilter(commits, e.opts.CommitFilters, e.opts.NoCaseSensitive)
	fmt.Printf("DEBUG: After filtering: %d commits\n", len(filteredCommits))

	for i, commit := range commits {
		fmt.Printf("DEBUG: Processing commit[%d]: hash=%s, type=%s, scope=%s, subject=%s\n",
			i, commit.Hash.Short, commit.Type, commit.Scope, commit.Subject)

		if commit.Merge != nil {
			fmt.Printf("DEBUG: Commit[%d] is a merge commit\n", i)
			mergeCommits = append(mergeCommits, commit)
			continue
		}

		if commit.Revert != nil {
			fmt.Printf("DEBUG: Commit[%d] is a revert commit\n", i)
			revertCommits = append(revertCommits, commit)
			continue
		}
	}

	for i, commit := range filteredCommits {
		if commit.Merge == nil && commit.Revert == nil {
			fmt.Printf("DEBUG: Processing commit[%d] for grouping: hash=%s, type=%s, scope=%s, subject=%s\n",
				i, commit.Hash.Short, commit.Type, commit.Scope, commit.Subject)

			raw, ttl := e.commitGroupTitle(commit)
			fmt.Printf("DEBUG: Commit[%d] group title: raw=%q, title=%q\n", i, raw, ttl)

			e.processCommitGroups(&commitGroups, commit, e.opts.NoCaseSensitive)
		}

		e.processNoteGroups(&noteGroups, commit)
	}

	e.sortCommitGroups(commitGroups)
	e.sortNoteGroups(noteGroups)

	fmt.Printf("DEBUG: Final result: %d commitGroups, %d mergeCommits, %d revertCommits, %d noteGroups\n",
		len(commitGroups), len(mergeCommits), len(revertCommits), len(noteGroups))

	// Debug: Zeige die erstellten Commit-Gruppen
	for i, group := range commitGroups {
		fmt.Printf("DEBUG: CommitGroup[%d]: title=%q, rawTitle=%q, commits=%d\n",
			i, group.Title, group.RawTitle, len(group.Commits))

		for j, commit := range group.Commits {
			fmt.Printf("DEBUG: CommitGroup[%d] commit[%d]: hash=%s, subject=%s\n",
				i, j, commit.Hash.Short, commit.Subject)
		}
	}

	return commitGroups, mergeCommits, revertCommits, noteGroups
}

func (e *commitExtractor) processCommitGroups(groups *[]*CommitGroup, commit *Commit, noCaseSensitive bool) {
	var group *CommitGroup

	// commit group
	raw, ttl := e.commitGroupTitle(commit)

	for _, g := range *groups {
		rawTitleTmp := g.RawTitle
		if noCaseSensitive {
			rawTitleTmp = strings.ToLower(g.RawTitle)
		}

		rawTmp := raw
		if noCaseSensitive {
			rawTmp = strings.ToLower(raw)
		}
		if rawTitleTmp == rawTmp {
			group = g
		}
	}

	if group != nil {
		group.Commits = append(group.Commits, commit)
	} else if raw != "" {
		*groups = append(*groups, &CommitGroup{
			RawTitle: raw,
			Title:    ttl,
			Commits:  []*Commit{commit},
		})
	}
}

func (e *commitExtractor) processNoteGroups(groups *[]*NoteGroup, commit *Commit) {
	if len(commit.Notes) != 0 {
		for _, note := range commit.Notes {
			e.appendNoteToNoteGroups(groups, note)
		}
	}
}

func (e *commitExtractor) appendNoteToNoteGroups(groups *[]*NoteGroup, note *Note) {
	exist := false

	for _, g := range *groups {
		if g.Title == note.Title {
			exist = true
			g.Notes = append(g.Notes, note)
		}
	}

	if !exist {
		*groups = append(*groups, &NoteGroup{
			Title: note.Title,
			Notes: []*Note{note},
		})
	}
}

func (e *commitExtractor) commitGroupTitle(commit *Commit) (string, string) {
	var (
		raw string
		ttl string
	)

	fmt.Printf("DEBUG: commitGroupTitle called for commit: hash=%s, type=%s, scope=%s, subject=%s\n",
		commit.Hash.Short, commit.Type, commit.Scope, commit.Subject)
	fmt.Printf("DEBUG: CommitGroupBy: %s\n", e.opts.CommitGroupBy)

	// Zeige die gesamte Commit-Struktur
	fmt.Printf("DEBUG: Commit structure: %+v\n", commit)

	if title, ok := dotGet(commit, e.opts.CommitGroupBy); ok {
		fmt.Printf("DEBUG: dotGet(%s) returned title: %v, ok: %v\n", e.opts.CommitGroupBy, title, ok)
		if v, ok := title.(string); ok {
			raw = v
			fmt.Printf("DEBUG: Raw title: %q\n", raw)
			if t, ok := e.opts.CommitGroupTitleMaps[v]; ok {
				ttl = t
				fmt.Printf("DEBUG: Mapped title: %q\n", ttl)
			} else {
				//nolint:staticcheck
				ttl = strings.Title(raw)
				fmt.Printf("DEBUG: Title-cased title: %q\n", ttl)
			}
		} else {
			fmt.Printf("DEBUG: title is not a string: %T\n", title)
		}
	} else {
		fmt.Printf("DEBUG: dotGet(%s) returned ok: false\n", e.opts.CommitGroupBy)
	}

	fmt.Printf("DEBUG: commitGroupTitle returning raw=%q, ttl=%q\n", raw, ttl)
	return raw, ttl
}

func (e *commitExtractor) sortCommitGroups(groups []*CommitGroup) { //nolint:gocyclo
	// NOTE(khos2ow): this function is over our cyclomatic complexity goal.
	// Be wary when adding branches, and look for functionality that could
	// be reasonably moved into an injected dependency.

	order := make(map[string]int)
	if e.opts.CommitGroupSortBy == "Custom" {
		for i, t := range e.opts.CommitGroupTitleOrder {
			order[t] = i
		}
	}

	// groups
	// TODO(khos2ow): move the inline sort function to
	// conceret implementation of sort.Interface in order
	// to reduce cyclomatic complaxity.
	sort.Slice(groups, func(i, j int) bool {
		if e.opts.CommitGroupSortBy == "Custom" {
			return order[groups[i].RawTitle] < order[groups[j].RawTitle]
		}

		var (
			a, b interface{}
			ok   bool
		)

		a, ok = dotGet(groups[i], e.opts.CommitGroupSortBy)
		if !ok {
			return false
		}

		b, ok = dotGet(groups[j], e.opts.CommitGroupSortBy)
		if !ok {
			return false
		}

		res, err := compare(a, "<", b)
		if err != nil {
			return false
		}
		return res
	})

	// commits
	for _, group := range groups {
		group := group // pin group to avoid potential bugs with passing group to lower functions

		// TODO(khos2ow): move the inline sort function to
		// conceret implementation of sort.Interface in order
		// to reduce cyclomatic complaxity.
		sort.Slice(group.Commits, func(i, j int) bool {
			var (
				a, b interface{}
				ok   bool
			)

			a, ok = dotGet(group.Commits[i], e.opts.CommitSortBy)
			if !ok {
				return false
			}

			b, ok = dotGet(group.Commits[j], e.opts.CommitSortBy)
			if !ok {
				return false
			}

			res, err := compare(a, "<", b)
			if err != nil {
				return false
			}
			return res
		})
	}
}

func (e *commitExtractor) sortNoteGroups(groups []*NoteGroup) {
	// groups
	sort.Slice(groups, func(i, j int) bool {
		return strings.ToLower(groups[i].Title) < strings.ToLower(groups[j].Title)
	})

	// notes
	for _, group := range groups {
		group := group // pin group to avoid potential bugs with passing group to lower functions
		sort.Slice(group.Notes, func(i, j int) bool {
			return strings.ToLower(group.Notes[i].Title) < strings.ToLower(group.Notes[j].Title)
		})
	}
}
