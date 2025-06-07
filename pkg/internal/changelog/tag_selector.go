package changelog

import (
	"fmt"
	"strings"
)

type tagSelector struct{}

func newTagSelector() *tagSelector {
	return &tagSelector{}
}

func (s *tagSelector) Select(tags []*Tag, query string) ([]*Tag, string, error) {
	if len(tags) == 0 {
		return nil, "", nil
	}

	// Debug: Print tags we're working with
	fmt.Printf("DEBUG: Tag selector input: %d tags, query=%s\n", len(tags), query)
	for i, tag := range tags {
		fmt.Printf("DEBUG: Input tag[%d]: %s\n", i, tag.Name)
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

		fmt.Printf("DEBUG: Parsed range query: from=%s, to=%s\n", from, to)

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

	fmt.Printf("DEBUG: selectRange: fromTag=%s, toTag=%s\n", fromTag.Name, toTag.Name)

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

	fmt.Printf("DEBUG: selectRange result: %d tags, first=%s\n", len(result), first)
	for i, tag := range result {
		fmt.Printf("DEBUG: Result tag[%d]: %s\n", i, tag.Name)
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

	fmt.Printf("DEBUG: selectFrom: fromTag=%s\n", fromTag.Name)

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

	fmt.Printf("DEBUG: selectFrom result: %d tags\n", len(result))
	for i, tag := range result {
		fmt.Printf("DEBUG: Result tag[%d]: %s\n", i, tag.Name)
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
		return res, "", errNotFoundTag
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
		return nil, "", errNotFoundTag
	}

	return res, from, nil
}
