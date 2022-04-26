package main

import (
	"errors"
	"fmt"
	"os"
	"path"
	"reflect"
	"testing"
)

func TestParseNote(t *testing.T) {
	tests := []struct {
		name        string
		fileName    string
		headerOnly  bool
		want        *Note
		wantedError error
	}{
		{
			name:        "simple with contents",
			fileName:    "my_test_diary.txt",
			headerOnly:  false,
			wantedError: nil,
			want: &Note{
				Path:  joinPath("my_test_diary.txt"),
				Title: "my_test_diary",
				Tags:  []string{"secret", "plzdontlook", "test"},
				Content: `
2022-04-23:

Dear Diary,

Today, I started adding unit tests to the program that created you.
I promise to look after you, and do you no harm by introducing bugs that might truncate your contents, leaving you empty inside.

I look forward to working with you.
`,
				rawHeader: `title: my_test_diary
tags:secret,plzdontlook,test
`,
			},
		},
		{
			name:        "ascii art in raw header",
			fileName:    "mic_drop.txt",
			headerOnly:  true,
			wantedError: nil,
			want: &Note{
				Path:    joinPath("mic_drop.txt"),
				Title:   "mic_drop",
				Tags:    []string{"DROPTHEMIC"},
				Content: "",
				rawHeader: `title: mic_drop
tags: DROPTHEMIC
mic:        @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
drop:       @@@@@@@@@@@@@@@@@@...,,,,,,,*@
commencing: @@@@@@@@@@@@@@@@@,,,,..,,,,,**
in:         @@@@@@@@@@@@@@@@..,,,,,...,***
5...:       @@@@@@@@@@@@@@((((,..,,,,*,,*@
4...:       @@@@@@@@@@@@(((((((((..****@@@
3...:       @@@@@@@@@@(((,,..(((((@@@@@@@@
2...:       @@@@@@@,((((((/((((@@@@@@@@@@@
1...:       @@@@@(((((((((((@@@@@@@@@@@@@@
DROPPING:   @@@%(((((((((@@@@@@@@@@@@@@@@@
THE:        @%%%%%((((,@@@@@@@@@@@@@@@@@@@
<howling>:  @@@%%%%%@@@@@@@@@@@@@@@@@@@@@@
`,
			},
		},
		{
			name:        "exclamation tag",
			fileName:    "another_one_rides_the_bus.txt",
			headerOnly:  true,
			wantedError: nil,
			want: &Note{
				Path:    joinPath("another_one_rides_the_bus.txt"),
				Title:   "another_one_rides_the_bus",
				Tags:    []string{"bumbumbum", "POW", "AnotherOneRidesTheBus", "!"},
				Content: "",
				rawHeader: `title: another_one_rides_the_bus
tags: bumbumbum, POW, AnotherOneRidesTheBus, !
`,
			},
		},
		{
			name:        "basic content check",
			fileName:    "basic.txt",
			headerOnly:  false,
			wantedError: nil,
			want: &Note{
				Path:    joinPath("basic.txt"),
				Title:   "lorem",
				Tags:    []string{"ipsum"},
				Content: "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.\n",
				rawHeader: `title: lorem
tags: ipsum
`,
			},
		},
		{
			name:        "empty header",
			fileName:    "bad_header_01.txt",
			headerOnly:  false,
			wantedError: ErrEmptyHeader,
			want:        nil,
		},
		{
			name:        "invalid header field",
			fileName:    "bad_header_02.txt",
			headerOnly:  false,
			wantedError: ErrInvalidHeader{line: "boop"},
			want:        nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseNoteConbini(tt.fileName, tt.headerOnly)

			if err != nil && tt.wantedError == nil {
				t.Fatalf("unexpected error occured while parsing note: %v", err)
			}

			if err == nil && tt.wantedError != nil {
				t.Fatalf("expected error, but did not get one")
			}

			if err != nil && tt.wantedError != nil && errors.Is(err, tt.wantedError) {
				if got != nil {
					t.Fatalf("successfully got error, but also got a result (expected nil)")
				}
				return
			} else if err != nil && tt.wantedError != nil && !errors.Is(err, tt.wantedError) {
				t.Fatalf("error mismatch wanted: %v, but got %v", tt.wantedError, err)
			}

			if tt.want == nil && got != nil {
				t.Fatalf("wanted nil result, but got a note")
			}
			if tt.want != nil && got == nil {
				t.Fatalf("wanted note, but got nil")
			}

			if tt.want.Path != got.Path {
				t.Errorf("Path mismatch:\nexpected: %v\ngot: %v", tt.want.Path, got.Path)
			}
			if tt.want.Title != got.Title {
				t.Errorf("Title mismatch:\nexpected: %v\ngot: %v", tt.want.Title, got.Title)
			}
			if !reflect.DeepEqual(tt.want.Tags, got.Tags) {
				t.Errorf("Tags mismatch:\nexpected: %v\ngot: %v", tt.want.Tags, got.Tags)
			}
			if tt.want.Content != got.Content {
				t.Errorf("Content mismatch:\nexpected: %+v\ngot: %+v", tt.want.Content, got.Content)
			}
			if tt.want.rawHeader != got.rawHeader {
				t.Errorf("rawHeader mismatch:\nexpected: %v\ngot: %v", tt.want.rawHeader, got.rawHeader)
			}
		})
	}
}

func joinPath(name string) string {
	// this can fail, but we're testing, so hopefully we've set up the required files before hand..
	curDir, err := os.Getwd()
	if err != nil {
		return fmt.Sprintf("couldn't get cwd: %v", err)
	}
	path := path.Join(curDir, "test_notes", name)
	return path
}

// Just a convenience function wrapped around ParseNote
func parseNoteConbini(name string, headerOnly bool) (*Note, error) {
	curDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("could not get working directory: %w", err)
	}
	path := path.Join(curDir, "test_notes", name)
	if !exists(path) {
		return nil, fmt.Errorf("file `%v` does not exist", path)
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("could not open file: %v, %w", path, err)
	}
	defer file.Close()

	return ParseNote(file, path, headerOnly)
}
