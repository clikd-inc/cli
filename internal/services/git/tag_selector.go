package git

import (
	"fmt"
	"strings"

	"clikd/internal/utils"
)

type tagSelector struct {
	logger utils.Logger
}

func newTagSelector(logger utils.Logger) *tagSelector {
	return &tagSelector{
		logger: logger,
	}
}

func (s *tagSelector) Select(tags []*Tag, query string) ([]*Tag, string, error) {
	if len(tags) == 0 {
		return nil, "", nil
	}

	// Debug: Print tags we're working with
	s.logger.Debug("Tag selector input", "tagCount", len(tags), "query", query)
	for i, tag := range tags {
		s.logger.Debug("Input tag", "index", i, "name", tag.Name)
	}

	// "<old>..<new>" pattern
	if strings.Contains(query, "..") {
		var (
			from = ""
			to   = ""
		)

		if query == ".." {
			return tags, "", nil
		}

		tokens := strings.Split(query, "..")
		if tokens[0] != "" {
			from = tokens[0]
		}
		if len(tokens) > 1 && tokens[1] != "" {
			to = tokens[1]
		}

		s.logger.Debug("Parsed range query", "from", from, "to", to)

		if from != "" && to != "" {
			return s.selectRange(tags, from, to)
		}

		if from != "" {
			return s.selectFrom(tags, from)
		}

		if to != "" {
			return s.selectTo(tags, to)
		}
	}

	// Select by tag name
	if tags, first, err := s.selectByTag(tags, query); err != nil {
		return nil, "", err
	} else if len(tags) != 0 {
		return tags, first, nil
	}

	// Fallback to default
	return tags, "", nil
}

func (s *tagSelector) selectRange(tags []*Tag, from, to string) ([]*Tag, string, error) {
	var (
		fromTag *Tag
		toTag   *Tag
		result  []*Tag
		first   string
	)

	// Find tag
	for _, tag := range tags {
		if tag.Name == to {
			toTag = tag
			continue
		}
		if tag.Name == from {
			fromTag = tag
			continue
		}
	}

	// Tag not found
	if toTag == nil {
		return nil, "", fmt.Errorf("\"%s\" tag is not found", to)
	}
	if fromTag == nil {
		return nil, "", fmt.Errorf("\"%s\" tag is not found", from)
	}

	s.logger.Debug("selectRange", "fromTag", fromTag.Name, "toTag", toTag.Name)

	// Find the range of tags
	var (
		inRange      = false
		foundOneItem = false
	)

	for _, tag := range tags {
		if tag.Name == fromTag.Name {
			inRange = true
		}

		if inRange {
			result = append(result, tag)
			foundOneItem = true
		}

		if tag.Name == toTag.Name && inRange {
			break
		}
	}

	if !foundOneItem {
		return nil, "", fmt.Errorf("we could not find any relevant tags")
	}

	first = from

	s.logger.Debug("selectRange result", "tagCount", len(result), "first", first)
	for i, tag := range result {
		s.logger.Debug("Result tag", "index", i, "name", tag.Name)
	}

	return result, first, nil
}

func (s *tagSelector) selectFrom(tags []*Tag, from string) ([]*Tag, string, error) {
	var (
		fromTag *Tag
		result  []*Tag
	)

	// Find tag
	for _, tag := range tags {
		if tag.Name == from {
			fromTag = tag
			break
		}
	}

	// Tag not found
	if fromTag == nil {
		return nil, "", fmt.Errorf("\"%s\" tag is not found", from)
	}

	s.logger.Debug("selectFrom", "fromTag", fromTag.Name)

	// Find the range of tags
	var (
		inRange      = false
		foundOneItem = false
	)

	for _, tag := range tags {
		if tag.Name == fromTag.Name {
			inRange = true
		}

		if inRange {
			result = append(result, tag)
			foundOneItem = true
		}
	}

	if !foundOneItem {
		return nil, "", fmt.Errorf("we could not find any relevant tags")
	}

	s.logger.Debug("selectFrom result", "tagCount", len(result))
	for i, tag := range result {
		s.logger.Debug("Result tag", "index", i, "name", tag.Name)
	}

	return result, from, nil
}

func (s *tagSelector) selectTo(tags []*Tag, to string) ([]*Tag, string, error) {
	var (
		res    []*Tag
		from   string
		enable bool
	)

	for i, tag := range tags {
		if tag.Name == to {
			enable = true
		}

		if enable {
			res = append(res, tag)
			from = ""
			if i+1 < len(tags) {
				from = tags[i+1].Name
			}
		}
	}

	if len(res) == 0 {
		return res, "", ErrNotFoundTag
	}

	return res, from, nil
}

func (s *tagSelector) selectByTag(tags []*Tag, query string) ([]*Tag, string, error) {
	var (
		res    []*Tag
		from   string
		enable bool
	)

	for i, tag := range tags {
		if tag.Name == query {
			enable = true
		}

		if enable {
			res = append(res, tag)
			from = ""
			if i+1 < len(tags) {
				from = tags[i+1].Name
			}
		}
	}

	if len(res) == 0 {
		return nil, "", ErrNotFoundTag
	}

	return res, from, nil
}
