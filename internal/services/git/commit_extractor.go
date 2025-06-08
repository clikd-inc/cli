package git

import (
	"fmt"
	"sort"
	"strings"

	"clikd/internal/utils"
)

type commitExtractor struct {
	opts   *Options
	logger utils.Logger
}

func newCommitExtractor(opts *Options, logger utils.Logger) *commitExtractor {
	if logger == nil {
		logger = utils.NewLogger("info", true).WithFields(map[string]interface{}{"module": "git"})
	}

	return &commitExtractor{
		opts:   opts,
		logger: logger,
	}
}

func (e *commitExtractor) Extract(commits []*Commit) ([]*CommitGroup, []*Commit, []*Commit, []*NoteGroup) {
	e.logger.Debug("commitExtractor.Extract called", "commits", len(commits))
	e.logger.Debug("Extractor configuration",
		"CommitGroupBy", e.opts.CommitGroupBy,
		"CommitSortBy", e.opts.CommitSortBy,
		"CommitGroupSortBy", e.opts.CommitGroupSortBy)
	e.logger.Debug("CommitGroupTitleMaps", "maps", e.opts.CommitGroupTitleMaps)

	commitGroups := []*CommitGroup{}
	noteGroups := []*NoteGroup{}
	mergeCommits := []*Commit{}
	revertCommits := []*Commit{}

	filteredCommits := commitFilter(commits, e.opts.CommitFilters, e.opts.NoCaseSensitive)
	e.logger.Debug("After filtering", "commits", len(filteredCommits))

	for i, commit := range commits {
		e.logger.Debug("Processing commit",
			"index", i,
			"hash", commit.Hash.Short,
			"type", commit.Type,
			"scope", commit.Scope,
			"subject", commit.Subject)

		if commit.Merge != nil {
			e.logger.Debug("Commit is a merge commit", "index", i)
			mergeCommits = append(mergeCommits, commit)
			continue
		}

		if commit.Revert != nil {
			e.logger.Debug("Commit is a revert commit", "index", i)
			revertCommits = append(revertCommits, commit)
			continue
		}
	}

	for i, commit := range filteredCommits {
		if commit.Merge == nil && commit.Revert == nil {
			e.logger.Debug("Processing commit for grouping",
				"index", i,
				"hash", commit.Hash.Short,
				"type", commit.Type,
				"scope", commit.Scope,
				"subject", commit.Subject)

			raw, ttl := e.commitGroupTitle(commit)
			e.logger.Debug("Commit group title", "index", i, "raw", raw, "title", ttl)

			e.processCommitGroups(&commitGroups, commit, e.opts.NoCaseSensitive)
		}

		e.processNoteGroups(&noteGroups, commit)
	}

	e.sortCommitGroups(commitGroups)
	e.sortNoteGroups(noteGroups)

	e.logger.Debug("Final result",
		"commitGroups", len(commitGroups),
		"mergeCommits", len(mergeCommits),
		"revertCommits", len(revertCommits),
		"noteGroups", len(noteGroups))

	// Debug: Zeige die erstellten Commit-Gruppen
	for i, group := range commitGroups {
		e.logger.Debug("CommitGroup",
			"index", i,
			"title", group.Title,
			"rawTitle", group.RawTitle,
			"commits", len(group.Commits))

		for j, commit := range group.Commits {
			e.logger.Debug("CommitGroup commit",
				"groupIndex", i,
				"commitIndex", j,
				"hash", commit.Hash.Short,
				"subject", commit.Subject)
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

	e.logger.Debug("commitGroupTitle called",
		"hash", commit.Hash.Short,
		"type", commit.Type,
		"scope", commit.Scope,
		"subject", commit.Subject)
	e.logger.Debug("CommitGroupBy", "value", e.opts.CommitGroupBy)

	// Zeige die gesamte Commit-Struktur
	e.logger.Debug("Commit structure", "commit", fmt.Sprintf("%+v", commit))

	if title, ok := utils.DotGet(commit, e.opts.CommitGroupBy); ok {
		e.logger.Debug("dotGet result", "field", e.opts.CommitGroupBy, "title", title, "ok", ok)
		if v, ok := title.(string); ok {
			raw = v
			e.logger.Debug("Raw title", "value", raw)
			if t, ok := e.opts.CommitGroupTitleMaps[v]; ok {
				ttl = t
				e.logger.Debug("Mapped title", "value", ttl)
			} else {
				//nolint:staticcheck
				ttl = strings.Title(raw)
				e.logger.Debug("Title-cased title", "value", ttl)
			}
		} else {
			e.logger.Debug("Title is not a string", "type", fmt.Sprintf("%T", title))
		}
	} else {
		e.logger.Debug("dotGet returned false", "field", e.opts.CommitGroupBy)
	}

	e.logger.Debug("commitGroupTitle returning", "raw", raw, "ttl", ttl)
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

		a, ok = utils.DotGet(groups[i], e.opts.CommitGroupSortBy)
		if !ok {
			return false
		}

		b, ok = utils.DotGet(groups[j], e.opts.CommitGroupSortBy)
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

			a, ok = utils.DotGet(group.Commits[i], e.opts.CommitSortBy)
			if !ok {
				return false
			}

			b, ok = utils.DotGet(group.Commits[j], e.opts.CommitSortBy)
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

// compare verwendet die zentrale Compare-Funktion
func compare(a interface{}, operator string, b interface{}) (bool, error) {
	return utils.Compare(a, operator, b)
}
