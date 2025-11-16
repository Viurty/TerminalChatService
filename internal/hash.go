package internal

import (
	"fmt"
	"log"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// Проверяем пароль + возвращаем роль
func CheckPassword(lines []string, name, password string) (bool, string) {
	hashPassword := ""
	role := ""

	for _, line := range lines {
		line = strings.TrimSpace(line)
		parts := strings.Split(line, ";")
		if parts[0] == name {
			hashPassword = parts[2]
			role = parts[1]
		}
	}

	err := bcrypt.CompareHashAndPassword([]byte(hashPassword), []byte(password))
	return err == nil, role
}

// Шифруем пароль
func encryptPassword(password string) string {
	hashPassword, _ := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(hashPassword)
}

// Записываем нового пользователя
func WritePassword(name, role, password, pathFile string) {
	file, err := os.OpenFile(pathFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("error during open file: %v", err)
		return
	}
	defer file.Close()

	hashPassword := encryptPassword(password)
	if _, err = file.WriteString(fmt.Sprintf("%s;%s;%s\n", name, role, hashPassword)); err != nil {
		log.Printf("error during write file: %v", err)
		return
	}
}
