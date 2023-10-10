package utils

import (
	"fmt"
	"os"
	"strings"

	"log/slog"
)

func LoadFromFile(filePath string, key []byte) string {
	fmt.Println("loading key")
	data, err := os.ReadFile(filePath)
	if err != nil {
		slog.Error(err.Error())
		return ""
	}
	if strings.HasPrefix(string(data), "0x") {
		fmt.Println("key unencrypted")
		data2 := strings.TrimPrefix(string(data), "0x")
		writeToFile(filePath, data2, key)

		return data2
	}
	fmt.Println("decrypting...")
	txtPlain, _ := Decrypt(string(data), key)
	return txtPlain
}

func writeToFile(filePath string, txt string, key []byte) {
	fmt.Println("storing file")
	txtEnc, _ := Encrypt(txt, key)
	data := []byte(txtEnc)
	err := os.WriteFile(filePath, data, 0644)
	if err != nil {
		slog.Error("Error writing file:" + err.Error())
	}
}
