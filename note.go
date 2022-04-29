package main

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"
)

const DIVIDER = "------"
const JOURNAL_DATE_FORMAT = "2006-01-02"

type Note struct {
	Path    string
	Title   string
	Tags    []string
	Content string
	// header may include metadata that's not necessarily tracked in this struct
	rawHeader string
}

//func (i Note) Title() string       { return i.Title }
func (i Note) Description() string { return strings.Join(i.Tags, ", ") }
func (i Note) FilterValue() string {
	return fmt.Sprintf("%s%s%s", i.Title, i.Path, strings.Join(i.Tags, ""))
}

type segmentType int

const (
	segmentTypeUnknown segmentType = iota
	segmentTypeEmpty
	segmentTypeDate
	segmentTypeTask
	segmentTypeText
	// segmentTypeBullet // TODO: figure out this one... for now maybe it'll just be part of text types
)

type NoteSegment struct {
	Type       segmentType
	content    string
	taskBullet string
}

func (ns NoteSegment) IsChecked() bool {
	return len(strings.TrimSpace(ns.taskBullet)) > 0
}

func (ns *NoteSegment) SetCheck(s string) {
	ns.taskBullet = s
}

func (ns NoteSegment) Content() string {
	if ns.Type == segmentTypeTask {
		_, content, _ := strings.Cut(ns.content, ":")
		return fmt.Sprintf("-[%v]:%v", ns.taskBullet, content)
	}
	return ns.content
}

var dateStampRE = regexp.MustCompile(`\s*?[0-9]{2,4}-[0-9]{1,2}-[0-9]{1,2}:`)
var taskRE = regexp.MustCompile(`\s*?-\s?\[([\s|\S]?)\]:.*`)

func Segmentize(content string) []NoteSegment {
	if len(content) == 0 {
		return []NoteSegment{
			{
				Type:    segmentTypeEmpty,
				content: "",
			},
		}
	}

	var lines []string
	sc := bufio.NewScanner(strings.NewReader(content))
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}

	segments := make([]NoteSegment, len(lines))
	for i, l := range lines {
		if len(l) == 0 || l == "\n" || l == "\r\n" {
			segments[i] = NoteSegment{
				Type:    segmentTypeEmpty,
				content: l,
			}
			continue
		}
		if dateStampRE.Match([]byte(l)) {
			segments[i] = NoteSegment{
				Type:    segmentTypeDate,
				content: l,
			}
			continue
		}
		if taskRE.Match([]byte(l)) {
			result := taskRE.FindStringSubmatch(l)
			taskBullet := ""
			if len(result) > 1 && len(result[1]) != 0 {
				taskBullet = result[1]
			}
			segments[i] = NoteSegment{
				Type:       segmentTypeTask,
				content:    l,
				taskBullet: taskBullet,
			}
			continue
		}

		// if nothing else matches it should just be plain text
		segments[i] = NoteSegment{
			Type:    segmentTypeText,
			content: l,
		}

	}
	// TODO?: group segment type. i.e. text segments adjacent to eachother can be put into a single group

	return segments
}

func Desegmentize(segments []NoteSegment) string {
	builder := strings.Builder{}
	if len(segments) == 0 || len(segments) == 1 && segments[0].Type == segmentTypeEmpty {
		return ""
	}

	for i, s := range segments {
		if s.Type != segmentTypeEmpty {
			builder.WriteString(fmt.Sprintf("%v", s.Content()))
			if i < len(segments)-1 {
				builder.WriteString("\n")
			}
		} else {
			builder.WriteString("\n")
		}
	}

	return builder.String()
}
