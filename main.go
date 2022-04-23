package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/urfave/cli/v2"
)

var build = "dev"

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

func ParseNote(reader io.Reader, path string, justHeader bool) (*Note, error) {
	in := bufio.NewReader(reader)
	var curLine string
	done := false
	isHeader := true
	buf := make([]byte, 2000)
	result := &Note{
		Path: path,
	}
	for !done {
		if isHeader {
			bytes, prefix, err := in.ReadLine()
			if errors.Is(err, io.EOF) {
				done = true
			}
			curLine += string(bytes)
			// prefix means it wasn't able to stick the full thing into the buffer
			if prefix {
				continue
			}
			if strings.TrimSpace(curLine) == DIVIDER {
				isHeader = false
				curLine = ""
				continue
			}
			result.rawHeader += fmt.Sprintf("%v\n", curLine)
			headerData := strings.Split(curLine, ":")
			if len(headerData) < 2 {
				fmt.Println(result.rawHeader)
				return nil, fmt.Errorf("could not parse header line: %v", curLine)
			}
			field := headerData[0]
			value := strings.Join(headerData[1:], ":")
			switch strings.TrimSpace(strings.ToLower(field)) {
			case "title":
				result.Title = strings.TrimSpace(value)
			case "tags":
				if len(value) == 0 {
					break
				}
				splitVal := strings.Split(value, ",")
				tags := make([]string, 0)
				for _, val := range splitVal {
					trimmed := strings.TrimSpace(val)
					if len(trimmed) == 0 {
						continue
					}
					contains := false
					for _, existingTag := range tags {
						if strings.EqualFold(existingTag, trimmed) {
							contains = true
						}
					}
					if !contains {
						tags = append(tags, trimmed)
					}
				}
				if len(tags) > 0 {
					result.Tags = tags
				}
			}
			curLine = ""
		} else if !justHeader {
			bytesRead, err := in.Read(buf)
			if errors.Is(err, io.EOF) {
				done = true
			}
			if bytesRead > 0 {
				curLine += string(buf[:bytesRead])
			}
		}
		if justHeader && !isHeader {
			done = true
		}

	}
	if !justHeader {
		result.Content = curLine
	}
	return result, nil
}

func addTimestamp(file *os.File, path string, ts time.Time) error {
	note, err := ParseNote(file, path, false)
	if err != nil {
		return err
	}

	note.Content = fmt.Sprintf("%v:\n\n\n%v", ts.Format(JOURNAL_DATE_FORMAT), note.Content)
	out := []byte(fmt.Sprintf("%v%v\n\n%v", note.rawHeader, DIVIDER, note.Content))
	file.Truncate(0)
	wrote := 0
	for written := 0; written <= len(out)-1; written += wrote {
		wrote, err = file.Write(out[written:])

		if err != nil {
			return fmt.Errorf("problem writing to file: %w", err)
		}
	}
	return nil
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}

func NewNoteFile(filePath string) error {
	if err := os.MkdirAll(filepath.Dir(filePath), 0770); err != nil {
		return err
	}

	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("could not create file (at path: %v): %w", filePath, err)
	}
	initialName := path.Base(filePath)

	fmt.Fprintf(f, `title: %v
tags:
%v
`, strings.TrimSuffix(initialName, ".txt"), DIVIDER,
	)
	defer f.Close()
	return nil
}

func checkExistance(userInput string) (string, error) {
	if !strings.HasSuffix(userInput, ".txt") {
		userInput += ".txt"
	}
	curDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("could not get working directory: %w", err)
	}
	outPath := path.Join(curDir, userInput)
	if exists(outPath) {
		return "", fmt.Errorf("file `%v` already exists", outPath)
	}
	return outPath, nil
}

func NewNote(c *cli.Context) error {
	name := c.Args().First()
	outPath, err := checkExistance(name)
	if err != nil {
		return err
	}
	return NewNoteFile(outPath)
}

func CheckNote(c *cli.Context) error {
	name := c.Args().First()
	outPath, err := checkExistance(name)
	if err != nil {
		return err
	}

	file, err := os.Open(outPath)
	if err != nil {
		return fmt.Errorf("could not open file: %v, %w", outPath, err)
	}
	defer file.Close()

	note, err := ParseNote(file, outPath, false)
	if err != nil {
		return fmt.Errorf("could not parse file: %w", err)
	}
	fmt.Printf("loaded file: \n「%v」\n%+q\n%v\n", note.Title, note.Tags, note.Content)

	return nil
}

func NewEntry(c *cli.Context) error {
	name := c.Args().First()
	outPath, err := checkExistance(name)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(outPath, os.O_RDWR|os.O_APPEND, os.ModePerm)
	if err != nil {
		return fmt.Errorf("could not open file: %v, %w", outPath, err)
	}
	defer file.Close()

	err = addTimestamp(file, outPath, time.Now())
	if err != nil {
		return fmt.Errorf("could not add timestamp to file: %w", err)
	}

	return nil
}

func CatNote(c *cli.Context) error {
	name := c.Args().First()
	outPath, err := checkExistance(name)
	if err != nil {
		return err
	}

	file, err := os.Open(outPath)
	if err != nil {
		return fmt.Errorf("could not open file: %v, %w", outPath, err)
	}
	defer file.Close()

	note, err := ParseNote(file, outPath, false)
	if err != nil {
		return fmt.Errorf("could not parse file: %w", err)
	}
	fmt.Printf("%v", note.Content)

	return nil
}

func files(dir string) []string {
	results := make([]string, 0)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(strings.ToLower(info.Name()), ".txt") {
			results = append(results, path)
		}
		return nil
	})
	if err != nil {
		return nil
	}
	return results
}

func checkTags(in chan string, out chan Note, wg *sync.WaitGroup, targetTag string) {
	for filename := range in {
		f, err := os.Open(filename)
		if err != nil {
			fmt.Printf("could not open file: %v", err)
			continue
		}
		note, err := ParseNote(f, filename, true)
		if err != nil {
			fmt.Printf("could not parse file: %v", err)
			continue
		}
		f.Close()

		for _, val := range note.Tags {
			if strings.Contains(strings.ToLower(val), strings.ToLower(targetTag)) {
				out <- *note
				continue
			}
		}

	}
	wg.Done()
}

func CheckTags(c *cli.Context) error {
	curDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not get working directory: %w", err)
	}
	fileList := files(curDir)
	wg := &sync.WaitGroup{}
	searchTag := c.Args().First()

	wg.Add(10)
	in := make(chan string, 10)
	out := make(chan Note, len(fileList))
	for i := 0; i < 10; i++ {
		go checkTags(in, out, wg, searchTag)
	}
	for _, fileName := range fileList {
		in <- fileName
	}
	close(in)

	wg.Wait()
	close(out)
	for result := range out {
		fmt.Printf("%v : %v\n", result.Title, result.Path)
	}

	return nil
}

func Version(_ *cli.Context) error {
	fmt.Printf("notes version %v\n", build)
	return nil
}

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:    "new",
				Aliases: []string{"n"},
				Usage:   "add a new note",
				Action:  NewNote,
			},
			{
				Name:   "check",
				Usage:  "parse check",
				Action: CheckNote,
			},
			{
				Name:   "cat",
				Usage:  "outputs the contents of the file without header",
				Action: CatNote,
			},
			{
				Name:   "tagged",
				Usage:  "find files with tag",
				Action: CheckTags,
			},
			{
				Name:    "entry",
				Aliases: []string{"e"},
				Usage:   "adds todays timestamp to the top of content",
				Action:  NewEntry,
			},
			{
				Name:   "version",
				Usage:  "current version",
				Action: Version,
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
