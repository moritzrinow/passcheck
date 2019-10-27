# passcheck
A simple command line tool, that helps you with checking your credentials.

## How does it work?

Passcheck stores your usernames and the SHA-512 hashes of your passwords in contexts, where each context is a key-value storage powered by 
https://github.com/syndtr/goleveldb. Contexts are supposed to represent some kind of platform, where you have existing credentials, e.g. Google or Facebook.

### Default data directories

#### Linux - /var/lib/passcheck
#### Windows - C:\ProgramData\passcheck

## Install

### Docker

##### Linux

docker build -t passcheck -f Linux.Dockerfile .

##### Windows

docker build -t passcheck -f Windows.Dockerfile .

### Compile it yourself

Make sure to have Make, Golang and Git (go get) installed.

Run in terminal:

make

## Usage

#### All flags are optional
#### [] - Mandatory argument
#### ([]) - Optional argument

### Adding credentials

#### passcheck add [context] [username] --f
#### -> Enter password
The --f flag forces an update if that entry already exists.

### Checking credentials

#### passcheck check [context] [username]
#### -> Enter password
Prints "CORRECT" if the passwords (hashes) matched or "INCORRECT" if invalid and let's you try again.

### Listing contexts & entries

#### passcheck list ([context]) --a
If context specified, it lists all usernames of the given context. If not it lists all contexts.
Specify the --a flag to retreive all contexts and their usernames.

### Removing contexts & entries

#### passcheck remove [context] ([username]) --f
If username is specified, it removes the specified entry from the context. If not it removes the whole context and it's entries.
Specify --f flag to bypass the prompt.

### Help

#### passcheck help
