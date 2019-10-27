package main

import (
	"runtime"
	"bufio"
	"bytes"
	"crypto/sha512"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/howeyc/gopass"
	"github.com/syndtr/goleveldb/leveldb"
)

type ParsedCommand struct {
	Name    string
	Flags   []string
	Args    []string
	Handler CommandHandler
}

type CommandDescription struct {
	Name          string
	MinArgCount   int
	PossibleFlags []string
	Handler       CommandHandler
	Help          string
}

func (c *CommandDescription) Print() {
	fmt.Println(c.Help)
}

// CommandHandler function type
type CommandHandler func(c *ParsedCommand)

var commands map[string]CommandDescription

func InitCommands() {
	commands = map[string]CommandDescription{
		"add":    {"add", 2, []string{"--f"}, HandleAdd, "add [context] [username] --f(overwrite)"},
		"check":  {"check", 2, []string{}, HandleCheck, "check [context] [usernane]"},
		"list":   {"list", 0, []string{"--a"}, HandleList, "list ([context]) --a(all contexts + usernames)"},
		"help":   {"help", 0, []string{}, HandleHelp, "help (lists commands)"},
		"remove": {"remove", 1, []string{"--f"}, HandleRemove, "remove [context] --f ([username])"},
	}
}

func Contains(s []string, str string) bool {
	for _, elem := range s {
		if elem == str {
			return true
		}
	}

	return false
}

func (p *ParsedCommand) FitsRules() (bool, error) {
	// check command name
	rules, ok := commands[p.Name]
	if !ok {
		return false, fmt.Errorf("Unknown command [%s]", p.Name)
	}

	// check args
	if len(p.Args) < rules.MinArgCount {
		return false, errors.New("Insufficient arguments")
	}

	// check flags
	for _, flag := range p.Flags {
		if !Contains(rules.PossibleFlags, flag) {
			return false, fmt.Errorf("Unknown flag %s", flag)
		}
	}

	p.Handler = rules.Handler

	return true, nil
}

func ParseCommandLineArgs(args []string) (*ParsedCommand, error) {
	if len(args) < 1 {
		return nil, errors.New("No command specified")
	}

	command := ParsedCommand{}
	command.Name = args[0]
	command.Handler = nil

	for i := 1; i < len(args); i++ {
		value := args[i]

		if strings.HasPrefix(value, "--") {
			command.Flags = append(command.Flags, value)
			continue
		}

		command.Args = append(command.Args, value)
	}

	return &command, nil
}

func HashEqual(hash1, hash2 []byte) bool {
	compare := bytes.Compare(hash1, hash2)
	return compare == 0
}

func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return true, err
}

func GetAnswer(message string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s[y/n]", message)
	input, err := reader.ReadString('\n')
	answer := string([]byte(input)[0])
	return err == nil && answer == "y"
}

func GetPassword() ([]byte, error) {
	fmt.Println("Enter password:")
	password, err := gopass.GetPasswdMasked()
	return password, err
}

func HashElements(context []byte, username []byte, password []byte) [64]byte {
	elements := append(append(context, username...), password...)
	return sha512.Sum512(elements)
}

func GetContexts() ([]string, error) {
	root := dataDir

	if ok, _ := Exists(root); !ok {
		return nil, nil
	}

	var directories []string

	files, err := ioutil.ReadDir(root)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() {
			directories = append(directories, file.Name())
		}
	}

	return directories, nil
}

func GetUserNames(context string) ([]string, error) {
	dbDir := fmt.Sprintf("%s/%s", dataDir, context)
	contextExists, _ := Exists(dbDir)
	if !contextExists {
		return nil, fmt.Errorf("Context:[%s] does not exist", context)
	}

	var usernames []string

	db, err := leveldb.OpenFile(dbDir, nil)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	iter := db.NewIterator(nil, nil)
	for iter.Next() {
		err = iter.Error()
		if err != nil {
			return nil, err
		}

		key := iter.Key()
		usernames = append(usernames, string(key))
	}
	iter.Release()

	return usernames, nil
}

func HandleList(p *ParsedCommand) {
	// No context specified
	if len(p.Args) < 1 {
		contexts, err := GetContexts()
		if err != nil {
			fmt.Println(err)
		}

		for _, context := range contexts {
			fmt.Println(context)

			if !Contains(p.Flags, "--a") {
				continue
			}

			// Also print usernames
			usernames, err := GetUserNames(context)
			if err != nil {
				fmt.Println(err)
			}

			for _, username := range usernames {
				fmt.Printf("-->%s\n", username)
			}
		}

		return
	}

	context := p.Args[0]
	usernames, err := GetUserNames(context)
	if err != nil {
		fmt.Println(err)
	}

	for _, username := range usernames {
		fmt.Println(username)
	}
}

func HandleHelp(p *ParsedCommand) {
	PrintCommands()
}

func HandleAdd(p *ParsedCommand) {
	contextName := p.Args[0]
	username := p.Args[1]

	// Ask for password safely
	password, err := GetPassword()
	if err != nil {
		fmt.Println(err)
		return
	}

	dbPath := fmt.Sprintf("%s/%s", dataDir, contextName)
	db, err := leveldb.OpenFile(dbPath, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	exists, err := db.Has([]byte(username), nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	if exists {
		if !Contains(p.Flags, "--f") {
			answer := GetAnswer("An entry with that username already exists in the specified context. Continue overwriting?")
			if !answer {
				return
			}
		}
	}

	hash := HashElements([]byte(contextName), []byte(username), password)
	err = db.Put([]byte(username), hash[:], nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Added entry for context:[%s], username:[%s]\n", contextName, username)
}

func HandleRemove(p *ParsedCommand) {
	contextName := p.Args[0]

	dbDir := fmt.Sprintf("%s/%s", dataDir, contextName)
	contextExists, _ := Exists(dbDir)
	if !contextExists {
		fmt.Printf("Context:[%s] does not exist.\n", contextName)
		return
	}

	// No username specified, so we delete whole context
	if len(p.Args) < 2 {
		if !Contains(p.Flags, "--f") {
			answer := GetAnswer("Continue to remove the context and all of it's entries?")
			if !answer {
				return
			}
		}

		err := os.RemoveAll(dbDir)
		if err != nil {
			fmt.Println(err)
			return
		}

		return
	}

	db, err := leveldb.OpenFile(dbDir, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer db.Close()

	username := p.Args[1]
	usernameBin := []byte(username)
	exists, err := db.Has(usernameBin, nil)
	if !exists {
		fmt.Printf("Could not delete entry for context:[%s], username:[%s]. Entry does not exist.\n", contextName, username)
		return
	}

	if !Contains(p.Flags, "--f") {
		prompt := GetAnswer("Continue to remove entry?")
		if !prompt {
			return
		}
	}

	err = db.Delete(usernameBin, nil)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("Successfully removed entry for context:[%s], username:[%s]\n", contextName, username)
}

func HandleCheck(p *ParsedCommand) {
	contextName := p.Args[0]

	dbDir := fmt.Sprintf("%s/%s", dataDir, contextName)
	contextExists, _ := Exists(dbDir)
	if !contextExists {
		fmt.Printf("Context:[%s] does not exist.\n", contextName)
		return
	}

	username := p.Args[1]
	db, err := leveldb.OpenFile(dbDir, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	exists, err := db.Has([]byte(username), nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	if !exists {
		fmt.Printf("No entry exists for context:[%s], username:[%s]\n", contextName, username)
		return
	}

	for {
		password, err := GetPassword()
		if err != nil {
			fmt.Println(err)
			return
		}

		hash := HashElements([]byte(contextName), []byte(username), password)
		actualHash, err := db.Get([]byte(username), nil)
		if err != nil {
			fmt.Println(err)
			return
		}

		if HashEqual(hash[:], actualHash) {
			fmt.Println("CORRECT!")
			return
		}

		fmt.Println("INCORRECT! Try again.")
	}
}

func PrintCommands() {
	for _, desc := range commands {
		desc.Print()
	}
}

var dataDir string

func Init() {
	InitCommands()

	switch runtime.GOOS {
	case "windows":
		dataDir = "C:\\ProgramData\\passcheck"
	case "linux":
		dataDir = "/var/lib/passcheck"
	default:
		dataDir = "./"
	}
}

func main() {
	Init()
	args := os.Args[1:]

	command, err := ParseCommandLineArgs(args)
	if err != nil {
		fmt.Println(err)
		PrintCommands()
		return
	}

	ok, err := command.FitsRules()
	if !ok && err != nil {
		fmt.Println(err)
		return
	}

	command.Handler(command)
}
