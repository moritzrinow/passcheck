# passcheck
A simple command line tool, that helps you with checking your credentials.

## How does it work?

Passcheck stores your usernames and the SHA-256 hashes of your passwords in contexts, where each context is a key-value storage powered by 
https://github.com/syndtr/goleveldb. Contexts are supposed to represent some kind of platform, where you have existing credentials, e.g. Google or Facebook.

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

## Commands

### add - Add an entry (or update)
### check - Check against an entry
### list - List contexts (and) usernames
### remove - Remove an entry or context
