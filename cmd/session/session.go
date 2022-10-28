package main

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math/rand"
	"os"
)

func main() {
	switch os.Args[1] {
	case "create":
		create()
	default:
		fmt.Printf("Invalid command: %v\n", os.Args[1])
	}
}

func create() {
	tokenBytes := make([]byte, 32)
	fmt.Printf("tokenBytes: %x\n", tokenBytes)
	nRead, err := rand.Read(tokenBytes)
	fmt.Printf("tokenBytes: %x\n", tokenBytes)
	fmt.Printf("nRead: %d\n", nRead)
	if err != nil {
		panic(fmt.Errorf("bytes: %w", err))
	}
	if nRead < 32 {
		panic(fmt.Errorf("bytes: didn't read enough random bytes"))
	}
	tokenHash := sha256.Sum256(tokenBytes)
	fmt.Printf("tokenHash: %x\n", tokenHash)
	tokenString := base64.URLEncoding.EncodeToString(tokenHash[:])
	fmt.Printf("tokenString: %s\n", tokenString)
}
