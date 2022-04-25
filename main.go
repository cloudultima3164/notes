package main

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

// These get overwritten at build time with goreleaser
var (
	build  = "dev"
	commit = "local"
	date   = "20XX-01-01"
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

func checkExistance(userInput string, wantExistance bool) (string, error) {
	if len(userInput) == 0 {
		return "", fmt.Errorf("empty filename")
	}
	if !strings.HasSuffix(userInput, ".txt") {
		userInput += ".txt"
	}
	curDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("could not get working directory: %w", err)
	}
	outPath := path.Join(curDir, userInput)
	if exists(outPath) != wantExistance {
		var message string
		if wantExistance {
			message = "file `%v` does not exist"
		} else {
			message = "file `%v` already exists"
		}
		return "", fmt.Errorf(message, outPath)
	}
	return outPath, nil
}

// TODO: hmmm maybe a validate files command?
//func CheckNote(c *cli.Context) error {
//	name := c.Args().First()
//	outPath, err := checkExistance(name, true)
//	if err != nil {
//		return err
//	}
//
//	file, err := os.Open(outPath)
//	if err != nil {
//		return fmt.Errorf("could not open file: %v, %w", outPath, err)
//	}
//	defer file.Close()
//
//	note, err := ParseNote(file, outPath, false)
//	if err != nil {
//		return fmt.Errorf("could not parse file: %w", err)
//	}
//	fmt.Printf("loaded file: \n「%v」\n%+q\n%v\n", note.Title, note.Tags, note.Content)
//
//	return nil
//}

func CatNote(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("could not open file: %v, %w", filePath, err)
	}
	defer file.Close()

	note, err := ParseNote(file, filePath, false)
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

func CheckTags(input []string) error {
	curDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not get working directory: %w", err)
	}
	fileList := files(curDir)
	wg := &sync.WaitGroup{}
	searchTag := strings.Join(input, " ")

	wg.Add(10)
	in := make(chan string, 10)
	out := make(chan Note, len(fileList))
	for i := 0; i < 10; i++ {
		go parseWorker(in, out, wg, true)
	}
	for _, fileName := range fileList {
		in <- fileName
	}
	close(in)

	wg.Wait()
	close(out)
	for result := range out {
		for _, val := range result.Tags {
			if strings.Contains(strings.ToLower(val), strings.ToLower(searchTag)) {
				fmt.Printf("%v : %v\n", result.Title, result.Path)
				break
			}
		}
	}

	return nil
}

func parseWorker(in chan string, out chan Note, wg *sync.WaitGroup, justHeader bool) {
	for filename := range in {
		f, err := os.Open(filename)
		if err != nil {
			fmt.Printf("could not open file: %v", err)
			continue
		}
		note, err := ParseNote(f, filename, justHeader)
		if err != nil {
			fmt.Printf("could not parse file: %v", err)
			continue
		}
		f.Close()

		out <- *note
	}
	wg.Done()
}

func collectFiles(justHeader bool) ([]Note, error) {
	curDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("could not get working directory: %w", err)
	}
	fileList := files(curDir)
	wg := &sync.WaitGroup{}

	wg.Add(10)
	in := make(chan string, 10)
	out := make(chan Note, len(fileList))
	for i := 0; i < 10; i++ {
		go parseWorker(in, out, wg, justHeader)
	}
	for _, fileName := range fileList {
		in <- fileName
	}
	close(in)

	wg.Wait()
	close(out)
	results := make([]Note, len(fileList))
	c := 0
	for result := range out {
		results[c] = result
		c++
	}

	return results, nil
}

var catCmd = &cobra.Command{
	Use:     "cat",
	Example: "notes cat [filepath]",
	Short:   "output the contents of a note",
	Long:    "output the contents of a note. if no note is specified, it goes into an interactive mode to select a note.",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return nil
		}
		preparedFileName, err := checkExistance(args[0], true)
		if err != nil {
			return err
		}
		args[0] = preparedFileName
		cmd.SetArgs(args)

		return nil
	},
	Run: func(_ *cobra.Command, args []string) {
		if len(args) == 0 {
			mod, err := NewFileSelector("Select File to Add an Entry to", true)
			if err != nil {
				fmt.Printf("Could not select a file: %v", err)
				return
			}
			m, err := tea.NewProgram(mod).StartReturningModel()
			if err != nil {
				fmt.Printf("Problem trying to get selection: %v", err)
				return
			}
			mod, ok := m.(model)
			if !ok {
				fmt.Println("Could not read selection")
				return
			}
			fmt.Printf("%s", mod.choice.Content)
			return

		}
		if err := CatNote(args[0]); err != nil {
			fmt.Printf("Problem trying to cat: %v", err)
		}
	},
}

var checkTagsCmd = &cobra.Command{
	Use:   "tagged",
	Short: "lists out files that match the tag",
	Args:  cobra.MinimumNArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		if err := CheckTags(args[0:]); err != nil {
			fmt.Printf("Problem trying to check for tags: %v", err)
		}
	},
}

var newEntryCmd = &cobra.Command{
	Use:     "entry",
	Aliases: []string{"e"},
	Short:   "adds today's YYYY-MM-DD date at the top of the provided note",
	Long:    "adds today's YYYY-MM-DD date at the top of the provided note. if no note is specified, it goes into an interactive mode to select one.",
	Example: "notes entry [directory/file]",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return nil
		}
		preparedFileName, err := checkExistance(args[0], true)
		if err != nil {
			return err
		}
		args[0] = preparedFileName
		cmd.SetArgs(args)

		return nil
	},
	Run: func(_ *cobra.Command, args []string) {
		var selectedFile string
		if len(args) == 0 {
			mod, err := NewFileSelector("Select File to Add a Date Entry to", true)
			if err != nil {
				fmt.Printf("Could not select a file: %v", err)
				return
			}
			m, err := tea.NewProgram(mod).StartReturningModel()
			if err != nil {
				fmt.Printf("Problem trying to get selection: %v", err)
				return
			}
			mod, ok := m.(model)
			if !ok {
				fmt.Println("Could not read selection")
				return
			}
			selectedFile = mod.choice.Path
		} else {
			selectedFile = args[0]
		}
		file, err := os.OpenFile(selectedFile, os.O_RDWR|os.O_APPEND, os.ModePerm)
		if err != nil {
			fmt.Printf("Could not open file: %v, %v", args[0], err)
			return
		}
		defer file.Close()

		err = addTimestamp(file, selectedFile, time.Now())
		if err != nil {
			fmt.Printf("Could not add timestamp to file: %v", err)
			return
		}
		fmt.Printf("Added %v entry line to %v", time.Now().Format(JOURNAL_DATE_FORMAT), selectedFile)

	},
}

var newNoteCmd = &cobra.Command{
	Use:     "new",
	Aliases: []string{"n"},
	Short:   "creates a new note at the given path/name",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("got an unexpected number of args (%v), expected %v", len(args), 1)
		}
		// TODO: if no filename is provided go into an interactive mode
		preparedFileName, err := checkExistance(args[0], false)
		if err != nil {
			return err
		}
		args[0] = preparedFileName
		cmd.SetArgs(args)

		return nil
	},
	Run: func(_ *cobra.Command, args []string) {
		if err := NewNoteFile(args[0]); err != nil {
			fmt.Printf("Problem trying to cat: %v", err)
		}
	},
}

var versionCmd = &cobra.Command{
	Use: "version",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("notes %v %v %v\n", build, commit, date)
	},
}

var rootCmd = &cobra.Command{
	Use:   "notes",
	Short: "Notes is a cli toolbox for plain text notes",
	Long: `A cli toolbox for creating and managing plain text notes. 
	all files are .txt so you do not need to specify .txt in the cli`,
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(checkTagsCmd)
	rootCmd.AddCommand(catCmd)
	rootCmd.AddCommand(newNoteCmd)
	rootCmd.AddCommand(newEntryCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	Execute()
}
