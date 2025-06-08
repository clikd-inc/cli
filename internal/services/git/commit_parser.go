package git

import (
	"clikd/internal/utils"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/tsuyoshiwada/go-gitcmd"
)

var (
	// constants
	separator = "@@__CHGLOG__@@"
	delimiter = "@@__CHGLOG_DELIMITER__@@"

	// fields
	hashField      = "HASH"
	authorField    = "AUTHOR"
	committerField = "COMMITTER"
	subjectField   = "SUBJECT"
	bodyField      = "BODY"

	// formats
	hashFormat      = hashField + ":%H\t%h"
	authorFormat    = authorField + ":%an\t%ae\t%at"
	committerFormat = committerField + ":%cn\t%ce\t%ct"
	subjectFormat   = subjectField + ":%s"
	bodyFormat      = bodyField + ":%b"

	// log
	logFormat = separator + strings.Join([]string{
		hashFormat,
		authorFormat,
		committerFormat,
		subjectFormat,
		bodyFormat,
	}, delimiter)
)

func joinAndQuoteMeta(list []string, sep string) string {
	return utils.JoinAndQuoteMeta(list, sep)
}

// Config enthält die Konfiguration für den Git-Service
type Config struct {
	Options *Options
}

// Options enthält die Optionen für den Git-Service
type Options struct {
	// Allgemeine Optionen
	NextTag               string              // Behandelt nicht freigegebene Commits als angegebene Tags (EXPERIMENTAL)
	TagFilterPattern      string              // Filtert Tag nach Regexp
	Sort                  string              // Gibt an, wie Tags sortiert werden; unterstützt derzeit "date" (Standard) oder "semver".
	NoCaseSensitive       bool                // Filtert Commits ohne Berücksichtigung von Groß- und Kleinschreibung
	CommitFilters         map[string][]string // Filtert durch Verwendung von `Commit`-Eigenschaften und -Werten. Filterung erfolgt nicht durch Angabe eines leeren Werts
	CommitSortBy          string              // Eigenschaftsname für Sortierung von `Commit` (z.B. `Scope`)
	CommitGroupBy         string              // Eigenschaftsname von `Commit`, der in `CommitGroup` gruppiert werden soll (z.B. `Type`)
	CommitGroupSortBy     string              // Eigenschaftsname für Sortierung von `CommitGroup` (z.B. `Title`)
	CommitGroupTitleOrder []string            // Vordefinierte sortierte Liste von Titeln für Sortierung von `CommitGroup`. Nur wenn `CommitGroupSortBy` ist `Custom`
	CommitGroupTitleMaps  map[string]string   // Zuordnung für Konvertierung von `CommitGroup`-Titeln

	// Parser-Optionen
	HeaderPattern     string   // Ein regulärer Ausdruck für das Parsen des Commit-Headers
	HeaderPatternMaps []string // Eine Regel für die Zuordnung des Ergebnisses von `HeaderPattern` zur Eigenschaft von `Commit`
	IssuePrefix       []string // Präfix für Issues (z.B. `#`, `gh-`)
	RefActions        []string // Wortliste von `Ref.Action`
	MergePattern      string   // Ein regulärer Ausdruck für das Parsen des Merge-Commits
	MergePatternMaps  []string // Ähnlich wie `HeaderPatternMaps`
	RevertPattern     string   // Ein regulärer Ausdruck für das Parsen des Revert-Commits
	RevertPatternMaps []string // Ähnlich wie `HeaderPatternMaps`
	NoteKeywords      []string // Schlüsselwortliste zum Finden von `Note`

	// Jira-Optionen
	JiraUsername                string            // Jira-Benutzername für API-Zugriff
	JiraToken                   string            // Jira-API-Token
	JiraURL                     string            // Jira-Server-URL
	JiraTypeMaps                map[string]string // Mapping von Jira-Issue-Typen zu Commit-Typen
	JiraIssueDescriptionPattern string            // Muster für Jira-Issue-Beschreibungen

	// Pfadfilterung
	Paths []string // Pfadfilter
}

type commitParser struct {
	logger                 utils.Logger
	client                 gitcmd.Client
	config                 *Config
	reHeader               *regexp.Regexp
	reMerge                *regexp.Regexp
	reRevert               *regexp.Regexp
	reRef                  *regexp.Regexp
	reIssue                *regexp.Regexp
	reNotes                *regexp.Regexp
	reMention              *regexp.Regexp
	reSignOff              *regexp.Regexp
	reCoAuthor             *regexp.Regexp
	reJiraIssueDescription *regexp.Regexp
}

func newCommitParser(logger utils.Logger, client gitcmd.Client, config *Config) *commitParser {
	opts := config.Options

	joinedRefActions := joinAndQuoteMeta(opts.RefActions, "|")
	joinedIssuePrefix := joinAndQuoteMeta(opts.IssuePrefix, "|")
	joinedNoteKeywords := joinAndQuoteMeta(opts.NoteKeywords, "|")

	return &commitParser{
		logger:                 logger,
		client:                 client,
		config:                 config,
		reHeader:               regexp.MustCompile(opts.HeaderPattern),
		reMerge:                regexp.MustCompile(opts.MergePattern),
		reRevert:               regexp.MustCompile(opts.RevertPattern),
		reRef:                  regexp.MustCompile("(?i)(" + joinedRefActions + ")\\s?([\\w/\\.\\-]+)?(?:" + joinedIssuePrefix + ")(\\d+)"),
		reIssue:                regexp.MustCompile("(?:" + joinedIssuePrefix + ")(\\d+)"),
		reNotes:                regexp.MustCompile("^(?i)\\s*(" + joinedNoteKeywords + ")[:\\s]+(.*)"),
		reMention:              regexp.MustCompile(`@([\w-]+)`),
		reSignOff:              regexp.MustCompile(`Signed-off-by:\s+([\p{L}\s\-\[\]]+)\s+<([\w+\-\[\].@]+)>`),
		reCoAuthor:             regexp.MustCompile(`Co-authored-by:\s+([\p{L}\s\-\[\]]+)\s+<([\w+\-\[\].@]+)>`),
		reJiraIssueDescription: regexp.MustCompile(opts.JiraIssueDescriptionPattern),
	}
}

func (p *commitParser) Parse(rev string) ([]*Commit, error) {
	p.logger.Debug("commitParser.Parse called", "rev", rev)

	paths := p.config.Options.Paths

	args := []string{
		rev,
		"--no-decorate",
		"--pretty=" + logFormat,
	}

	// Debug: Zeige die verwendeten Argumente
	p.logger.Debug("git log args", "args", args)

	// Debug: Zeige die Pfad-Filter, falls vorhanden
	if len(paths) > 0 {
		p.logger.Debug("Filtering commits by paths", "paths", paths)
		args = append(args, "--")
		args = append(args, paths...)
	}

	// Debug: Führe einen direkten Git-Befehl aus, um zu überprüfen, ob Commits vorhanden sind
	baseArgs := []string{"-C", "/Users/nyxb/Projects/nyxb/cli/clikd/test_repo", "log", rev, "--pretty=format:%h - %s"}
	p.logger.Debug("Executing direct git command", "command", "git "+strings.Join(baseArgs, " "))
	if out, err := exec.Command("git", baseArgs...).CombinedOutput(); err == nil {
		p.logger.Debug("Direct git log output", "output", string(out))
	} else {
		p.logger.Debug("Error in direct git log", "error", err, "output", string(out))
	}

	// Debug: Zeige das aktuelle Arbeitsverzeichnis
	if wd, err := os.Getwd(); err == nil {
		p.logger.Debug("Working directory", "path", wd)
	}

	p.logger.Debug("Executing git log with custom format")
	out, err := p.client.Exec("log", args...)

	if err != nil {
		p.logger.Error("Error in git log command", "error", err)
		return nil, err
	}

	// Debug: Zeige die ersten 200 Zeichen der Ausgabe
	if len(out) > 0 {
		preview := out
		if len(out) > 200 {
			preview = out[:200] + "..."
		}
		p.logger.Debug("Git log output preview", "output", preview)
	} else {
		p.logger.Debug("Git log output is empty")
	}

	lines := strings.Split(out, separator)

	// Debug: Zeige die Anzahl der Zeilen vor und nach dem Entfernen der ersten Zeile
	p.logger.Debug("Git log output split into lines", "count", len(lines))

	if len(lines) > 0 {
		lines = lines[1:]
		p.logger.Debug("After removing first line", "count", len(lines))
	} else {
		p.logger.Debug("No lines found in git log output")
	}

	commits := make([]*Commit, len(lines))

	for i, line := range lines {
		// Debug: Informationen zur Verarbeitung jeder Zeile
		p.logger.Debug("Processing line", "index", i, "length", len(line))

		commit := p.parseCommit(line)

		// Debug: Zeige Informationen zum geparsten Commit
		if commit != nil && commit.Hash != nil {
			p.logger.Debug("Parsed commit", "hash", commit.Hash.Short, "subject", commit.Subject)
		} else {
			p.logger.Debug("Failed to parse commit from line", "index", i)
		}

		commits[i] = commit
	}

	// Debug: Zeige die Anzahl der resultierenden Commits
	validCommits := 0
	for _, c := range commits {
		if c != nil {
			validCommits++
		}
	}
	p.logger.Debug("Parsed commits", "valid", validCommits, "total", len(lines))

	return commits, nil
}

func (p *commitParser) parseCommit(input string) *Commit {
	commit := &Commit{}
	tokens := strings.Split(input, delimiter)

	for _, token := range tokens {
		firstSep := strings.Index(token, ":")
		field := token[0:firstSep]
		value := strings.TrimSpace(token[firstSep+1:])

		switch field {
		case hashField:
			commit.Hash = p.parseHash(value)
		case authorField:
			commit.Author = p.parseAuthor(value)
		case committerField:
			commit.Committer = p.parseCommitter(value)
		case subjectField:
			p.processHeader(commit, value)
		case bodyField:
			p.processBody(commit, value)
		}
	}

	commit.Refs = p.uniqRefs(commit.Refs)
	commit.Mentions = p.uniqMentions(commit.Mentions)

	return commit
}

func (p *commitParser) parseHash(input string) *Hash {
	arr := strings.Split(input, "\t")

	return &Hash{
		Long:  arr[0],
		Short: arr[1],
	}
}

func (p *commitParser) parseAuthor(input string) *Author {
	arr := strings.Split(input, "\t")
	ts, err := strconv.Atoi(arr[2])
	if err != nil {
		ts = 0
	}

	return &Author{
		Name:  arr[0],
		Email: arr[1],
		Date:  time.Unix(int64(ts), 0),
	}
}

func (p *commitParser) parseCommitter(input string) *Committer {
	author := p.parseAuthor(input)

	return &Committer{
		Name:  author.Name,
		Email: author.Email,
		Date:  author.Date,
	}
}

func (p *commitParser) processHeader(commit *Commit, input string) {
	opts := p.config.Options

	// header (raw)
	commit.Header = input

	var res [][]string

	// Type, Scope, Subject etc ...
	res = p.reHeader.FindAllStringSubmatch(input, -1)
	if len(res) > 0 {
		p.logger.Debug("processHeader: Header matches found", "matches", res[0])
		p.logger.Debug("processHeader: HeaderPatternMaps", "maps", opts.HeaderPatternMaps)

		// Zeige, was mit den extrahierten Werten passieren soll
		if len(res[0]) > 1 {
			p.logger.Debug("processHeader: Extracted values", "values", res[0][1:])
			p.logger.Debug("processHeader: Will assign to fields", "fields", opts.HeaderPatternMaps)
		}

		utils.AssignDynamicValues(commit, opts.HeaderPatternMaps, res[0][1:])

		// Nach der Zuweisung zeigen, welche Werte gesetzt wurden
		p.logger.Debug("processHeader: After assignment",
			"Type", commit.Type,
			"Scope", commit.Scope,
			"Subject", commit.Subject)
	} else {
		p.logger.Debug("processHeader: No header matches found", "pattern", opts.HeaderPattern)
	}

	// Merge
	res = p.reMerge.FindAllStringSubmatch(input, -1)
	if len(res) > 0 {
		merge := &Merge{}
		utils.AssignDynamicValues(merge, opts.MergePatternMaps, res[0][1:])
		commit.Merge = merge
	}

	// Revert
	res = p.reRevert.FindAllStringSubmatch(input, -1)
	if len(res) > 0 {
		revert := &Revert{}
		utils.AssignDynamicValues(revert, opts.RevertPatternMaps, res[0][1:])
		commit.Revert = revert
	}

	// refs & mentions
	commit.Refs = p.parseRefs(input)
	commit.Mentions = p.parseMentions(input)
}

func (p *commitParser) extractLineMetadata(commit *Commit, line string) bool {
	meta := false

	refs := p.parseRefs(line)
	if len(refs) > 0 {
		meta = true
		commit.Refs = append(commit.Refs, refs...)
	}

	mentions := p.parseMentions(line)
	if len(mentions) > 0 {
		meta = true
		commit.Mentions = append(commit.Mentions, mentions...)
	}

	coAuthors := p.parseCoAuthors(line)
	if len(coAuthors) > 0 {
		meta = true
		commit.CoAuthors = append(commit.CoAuthors, coAuthors...)
	}

	signers := p.parseSigners(line)
	if len(signers) > 0 {
		meta = true
		commit.Signers = append(commit.Signers, signers...)
	}

	return meta
}

func (p *commitParser) processBody(commit *Commit, input string) {
	input = utils.ConvNewline(input, "\n")

	// body
	commit.Body = input

	// notes & refs & mentions
	commit.Notes = []*Note{}
	inNote := false
	trim := false
	fenceDetector := newMdFenceDetector()
	lines := strings.Split(input, "\n")

	// body without notes & refs & mentions
	trimmedBody := make([]string, 0, len(lines))

	for _, line := range lines {
		if !inNote {
			trim = false
		}
		fenceDetector.Update(line)

		if !fenceDetector.InCodeblock() && p.extractLineMetadata(commit, line) {
			trim = true
			inNote = false
		}
		// Q: should this check also only be outside of code blocks?
		res := p.reNotes.FindAllStringSubmatch(line, -1)

		if len(res) > 0 {
			inNote = true
			trim = true
			for _, r := range res {
				commit.Notes = append(commit.Notes, &Note{
					Title: r[1],
					Body:  r[2],
				})
			}
		} else if inNote {
			last := commit.Notes[len(commit.Notes)-1]
			last.Body = last.Body + "\n" + line
		}

		if !trim {
			trimmedBody = append(trimmedBody, line)
		}
	}

	commit.TrimmedBody = strings.TrimSpace(strings.Join(trimmedBody, "\n"))
	p.trimSpaceInNotes(commit)
}

func (*commitParser) trimSpaceInNotes(commit *Commit) {
	for _, note := range commit.Notes {
		note.Body = strings.TrimSpace(note.Body)
	}
}

func (p *commitParser) parseRefs(input string) []*Ref {
	refs := []*Ref{}

	// references
	res := p.reRef.FindAllStringSubmatch(input, -1)

	for _, r := range res {
		refs = append(refs, &Ref{
			Action: r[1],
			Source: r[2],
			Ref:    r[3],
		})
	}

	// issues
	res = p.reIssue.FindAllStringSubmatch(input, -1)
	for _, r := range res {
		duplicate := false
		for _, ref := range refs {
			if ref.Ref == r[1] {
				duplicate = true
			}
		}
		if !duplicate {
			refs = append(refs, &Ref{
				Action: "",
				Source: "",
				Ref:    r[1],
			})
		}
	}

	return refs
}

func (p *commitParser) parseSigners(input string) []Contact {
	res := p.reSignOff.FindAllStringSubmatch(input, -1)
	contacts := make([]Contact, len(res))

	for i, r := range res {
		contacts[i].Name = r[1]
		contacts[i].Email = r[2]
	}

	return contacts
}

func (p *commitParser) parseCoAuthors(input string) []Contact {
	res := p.reCoAuthor.FindAllStringSubmatch(input, -1)
	contacts := make([]Contact, len(res))

	for i, r := range res {
		contacts[i].Name = r[1]
		contacts[i].Email = r[2]
	}

	return contacts
}

func (p *commitParser) parseMentions(input string) []string {
	res := p.reMention.FindAllStringSubmatch(input, -1)
	mentions := make([]string, len(res))

	for i, r := range res {
		mentions[i] = r[1]
	}

	return mentions
}

func (p *commitParser) uniqRefs(refs []*Ref) []*Ref {
	arr := []*Ref{}

	for _, ref := range refs {
		exist := false
		for _, r := range arr {
			if ref.Ref == r.Ref && ref.Action == r.Action && ref.Source == r.Source {
				exist = true
			}
		}
		if !exist {
			arr = append(arr, ref)
		}
	}

	return arr
}

func (p *commitParser) uniqMentions(mentions []string) []string {
	arr := []string{}

	for _, mention := range mentions {
		exist := false
		for _, m := range arr {
			if mention == m {
				exist = true
			}
		}
		if !exist {
			arr = append(arr, mention)
		}
	}

	return arr
}

var (
	fenceTypes = []string{
		"```",
		"~~~",
		"    ",
		"\t",
	}
)

type mdFenceDetector struct {
	fence int
}

func newMdFenceDetector() *mdFenceDetector {
	return &mdFenceDetector{
		fence: -1,
	}
}

func (d *mdFenceDetector) InCodeblock() bool {
	return d.fence > -1
}

func (d *mdFenceDetector) Update(input string) {
	for i, s := range fenceTypes {
		if d.fence < 0 {
			if strings.Index(input, s) == 0 {
				d.fence = i
				break
			}
		} else {
			if strings.Index(input, s) == 0 && i == d.fence {
				d.fence = -1
				break
			}
		}
	}
}
