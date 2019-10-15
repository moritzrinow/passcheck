package main

import (
	"bufio"
	"bytes"
	"crypto/sha512"
	"fmt"
	"os"

	"github.com/howeyc/gopass"
	"github.com/syndtr/goleveldb/leveldb"
)

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
	password, err := gopass.GetPasswd()
	return password, err
}

func GetUsername() (string, error) {
	fmt.Println("Enter username:")
	var username string
	_, err := fmt.Scanf("%s\n", &username)
	return username, err
}

// GetContext gets the context name from Stdin
func GetContext() (string, error) {
	fmt.Println("Enter context name:")
	var contextName string
	_, err := fmt.Scanf("%s\n", &contextName)
	return contextName, err
}

func HashElements(context []byte, username []byte, password []byte) [64]byte {
	elements := append(append(context, username...), password...)
	return sha512.Sum512(elements)
}

func HandleAdd() {
	// Ask for passcheck context (like Google, Facebook, Gmail,...)
	contextName, err := GetContext()
	if err != nil {
		fmt.Println(err)
		return
	}

	// Ask for username
	username, err := GetUsername()
	if err != nil {
		fmt.Println(err)
		return
	}

	// Ask for password safely
	password, err := GetPassword()
	if err != nil {
		fmt.Println(err)
		return
	}

	dbPath := fmt.Sprintf("./data/%s", contextName)
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
		answer := GetAnswer("An entry with that username already exists in the specified context. Continue overwriting?")
		if !answer {
			return
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

func HandleCheck() {
	contextName, err := GetContext()
	if err != nil {
		fmt.Println(err)
		return
	}

	dbDir := fmt.Sprintf("./data/%s", contextName)
	contextExists, _ := Exists(dbDir)
	if !contextExists {
		fmt.Printf("Context:[%s] does not exist.\n", contextName)
		return
	}

	username, err := GetUsername()
	if err != nil {
		fmt.Println(err)
		return
	}

	db, err := leveldb.OpenFile(dbDir, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

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
	fmt.Println("List of passcheck commands:")
	fmt.Println("add\t\tAdd a new entry.")
	fmt.Println("check\t\tCheck a password against an entry.")
}

// CommandHandler function type
type CommandHandler func()

var commands = map[string]CommandHandler{
	"add":   HandleAdd,
	"check": HandleCheck,
}

func main() {
	args := os.Args[1:]

	if len(args) < 1 {
		fmt.Println("No command specified.")
		PrintCommands()
		return
	}

	command := args[0]

	handler, ok := commands[command]
	if !ok {
		fmt.Println("Command does not exist.")
		PrintCommands()
		return
	}
	handler()
}
