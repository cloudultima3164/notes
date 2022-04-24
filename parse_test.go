package main

import {
	"testing"
	"io"
	"os"
	"fmt"
	"cmp"
	"path"
	"errors"
}

func TestParse(t *testing.T) {
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
			return nil, fmt.Errorf("could not open file: %v, %w", outPath, err)
		}
		defer file.Close()

		note, err := ParseNote(file, path, false)
		if err != nil {
			return nil, fmt.Errorf("could not parse file: %v", err)
		}

		note, nil
	}
	

	tests := map[string]struct {
		name string
		headerOnly bool
		want Note
		
	} {
		"simple with contents": { name: "my_test_diary.txt", headerOnly: false, Note {  } }
	}

}
