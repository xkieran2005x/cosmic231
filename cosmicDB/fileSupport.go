package cosmicDB

import (
	"os"
	"log"
)

func fileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func deleteFile(name string) {
	err := os.Remove(name)
	if err != nil {
		log.Println(err)
		return
	}
}

func deleteFileIfExists(name string) {
	if fileExists(name) {deleteFile(name)}

}
