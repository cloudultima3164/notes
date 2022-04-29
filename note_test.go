package main

import (
	"reflect"
	"testing"
)

func TestSegmentizeAndDesegmentize(t *testing.T) {
	tests := []struct {
		name       string
		rawContent string
		segments   []NoteSegment
	}{
		{
			name:       "empty",
			rawContent: "",
			segments: []NoteSegment{
				{
					Type:    segmentTypeEmpty,
					content: "",
				},
			},
		},
		{
			name:       "multiple blank lines",
			rawContent: "\n\n",
			segments: []NoteSegment{
				{
					Type:    segmentTypeEmpty,
					content: "",
				},
				{
					Type:    segmentTypeEmpty,
					content: "",
				},
			},
		},
		{
			name:       "date with no space",
			rawContent: "2021-02-02:",
			segments: []NoteSegment{
				{
					Type:    segmentTypeDate,
					content: "2021-02-02:",
				},
			},
		},
		{
			name:       "date with space on sides space",
			rawContent: "   2021-02-02:   ",
			segments: []NoteSegment{
				{
					Type:    segmentTypeDate,
					content: "   2021-02-02:   ",
				},
			},
		},
		{
			name:       "date with text after",
			rawContent: "2021-02-02: some content here",
			segments: []NoteSegment{
				{
					Type:    segmentTypeDate,
					content: "2021-02-02: some content here",
				},
			},
		},
		{
			name:       "date with space before and text after",
			rawContent: "  2021-02-02: some content here",
			segments: []NoteSegment{
				{
					Type:    segmentTypeDate,
					content: "  2021-02-02: some content here",
				},
			},
		},
		{
			name:       "date and text",
			rawContent: "  2021-02-02:\nsome content here",
			segments: []NoteSegment{
				{
					Type:    segmentTypeDate,
					content: "  2021-02-02:",
				},
				{
					Type:    segmentTypeText,
					content: "some content here",
				},
			},
		},
		{
			name:       "multi text line break in middle",
			rawContent: "some content here\n\nothercontent",
			segments: []NoteSegment{
				{
					Type:    segmentTypeText,
					content: "some content here",
				},
				{
					Type:    segmentTypeEmpty,
					content: "",
				},
				{
					Type:    segmentTypeText,
					content: "othercontent",
				},
			},
		},
		{
			name:       "multi text line",
			rawContent: "some content here\nothercontent",
			segments: []NoteSegment{
				{
					Type:    segmentTypeText,
					content: "some content here",
				},
				{
					Type:    segmentTypeText,
					content: "othercontent",
				},
			},
		},
		{
			name:       "task",
			rawContent: "-[]: sometask",
			segments: []NoteSegment{
				{
					Type:    segmentTypeTask,
					content: "-[]: sometask",
				},
			},
		},
		{
			name:       "marked task",
			rawContent: "-[x]: sometask",
			segments: []NoteSegment{
				{
					Type:       segmentTypeTask,
					content:    "-[x]: sometask",
					taskBullet: "x",
				},
			},
		},
		{
			name:       "empty status task",
			rawContent: "-[ ]: sometask",
			segments: []NoteSegment{
				{
					Type:       segmentTypeTask,
					content:    "-[ ]: sometask",
					taskBullet: " ",
				},
			},
		},
		{
			name:       "symbol status task",
			rawContent: "-[ø]: sometask",
			segments: []NoteSegment{
				{
					Type:       segmentTypeTask,
					content:    "-[ø]: sometask",
					taskBullet: "ø",
				},
			},
		},
		{
			name:       "tasks",
			rawContent: "-[]: sometask\n-[]: another task",
			segments: []NoteSegment{
				{
					Type:    segmentTypeTask,
					content: "-[]: sometask",
				},
				{
					Type:    segmentTypeTask,
					content: "-[]: another task",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Segmentize(tt.rawContent); !reflect.DeepEqual(got, tt.segments) {
				t.Errorf("Segmentize() = %v, want %v", got, tt.segments)
			}

			if got := Desegmentize(tt.segments); !reflect.DeepEqual(got, tt.rawContent) {
				t.Errorf("Desegmentize() = %v, want %v", got, tt.rawContent)
			}
		})
	}
}
