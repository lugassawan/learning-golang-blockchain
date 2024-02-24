package utils

import (
	"fmt"
	"log"
	"os"
)

const dbFile = "./database/blockchain_%s.db"

func CheckDB(nodeId string) bool {
	if _, err := os.Stat(GetDBPath(nodeId)); os.IsNotExist(err) {
		return false
	}

	return true
}

func CreateDB(nodeId string) {
	if CheckDB(nodeId) {
		fmt.Printf("DB %s already exists\n", nodeId)
		return
	}

	file, err := os.Create(GetDBPath(nodeId))
	if err != nil {
		log.Panic(err)
	}

	defer file.Close()
	fmt.Printf("DB %s created successfully\n", nodeId)
}

func GetDBPath(nodeId string) string {
	return fmt.Sprintf(dbFile, nodeId)
}
