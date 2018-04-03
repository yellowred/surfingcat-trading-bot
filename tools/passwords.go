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
	//password = []byte("AAA")
	log.Println(string(password))
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic("Permissions: bcrypt password hashing unsuccessful")
	}
	log.Println(string(hash))
	if bcrypt.CompareHashAndPassword([]byte(""), password) != nil {
		log.Fatalln("bad password")
	}

}
