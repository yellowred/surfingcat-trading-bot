package main

import (
	"io/ioutil"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"
)

// How to use: `echo to_be_encrypted | go run tools/passwords.go
func main() {
	password, err := ioutil.ReadAll(os.Stdin)
	password = password[:len(password)-1] // trim EOF
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic("Permissions: bcrypt password hashing unsuccessful")
	}
	log.Println(string(hash))
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		log.Fatalln("bad password")
	}

}
