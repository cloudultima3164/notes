package main

import (
	"testing"
	"os"
	"fmt"
	"path"
	"reflect"
)

func TestParse(t *testing.T) {
	tests := map[string]struct {
		name string
		headerOnly bool
		want Note
		
	} {
		"simple with contents": { 
			name: "my_test_diary.txt",
			headerOnly: false,
			want: Note { 
				Path: joinPath("my_test_diary.txt"),
				Title: "my_test_diary",
				Tags: []string{ "secret", "plzdontlook", "test" },
				Content: 
`
2022-04-23:

Dear Diary,

Today, I started adding unit tests to the program that created you.
I promise to look after you, and do you no harm by introducing bugs that might truncate your contents, leaving you empty inside.

I look forward to working with you.
`,
				rawHeader: "",
			},
		},
		"ascii art in raw header": {
			name: "mic_drop.txt",
			headerOnly: true,
			want: Note { 
				Path: joinPath("mic_drop.txt"),
				Title: "mic_drop",
				Tags: []string{ "DROPTHEMIC" },
				Content: "",
				rawHeader:
`
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
		"exclamation tag": {
			name: "another_one_rides_the_bus.txt",
			headerOnly: true,
			want: Note { 
				Path: joinPath("another_one_rides_the_bus.txt"),
				Title: "another_one_rides_the_bus",
				Tags: []string{ "bumbumbum", "POW", "AnotherOneRidesTheBus", "!" },
				Content: "",
				rawHeader: "",
			},
		},
	}

	for name, struc := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := parseNoteConbini(struc.name, struc.headerOnly)
			if err != nil {
				t.Fatalf("error occured while parsing note: %v", err)
			}
			if !reflect.DeepEqual(struc.want, got) {
				t.Fatalf("expected: %v\ngot: %v", struc.want, *got)
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

	note, err := ParseNote(file, path, false)
	if err != nil {
		return nil, fmt.Errorf("could not parse file: %v", err)
	}

	return note, nil
}
