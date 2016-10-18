// Package htpasswd is utility package to manipulate htpasswd files. I supports\
// bcrypt and sha hashes.
package htpasswd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

// HashedPasswords name => hash
type HashedPasswords map[string]string

// HashAlgorithm enum for hashing algorithms
type HashAlgorithm string

const (
	// HashBCrypt bcrypt - recommended
	HashBCrypt = "bcrypt"
	// HashSHA sha5 insecure - do not use
	HashSHA = "sha"
)

const (
	// PasswordSeparator separates passwords from hashes
	PasswordSeparator = ":"
	// LineSeparator separates password records
	LineSeparator = "\n"
)

// MaxHtpasswdFilesize if your htpassd file is larger than 8MB, then your are doing it wrong
const MaxHtpasswdFilesize = 8 * 1024 * 1024 * 1024

// Bytes bytes representation
func (hp HashedPasswords) Bytes() (passwordBytes []byte) {
	passwordBytes = []byte{}
	for name, hash := range hp {
		passwordBytes = append(passwordBytes, []byte(name+PasswordSeparator+hash+LineSeparator)...)
	}
	return passwordBytes
}

// WriteToFile put them to a file will be overwritten or created
func (hp HashedPasswords) WriteToFile(file string) error {
	return ioutil.WriteFile(file, hp.Bytes(), 0644)
}

// SetPassword set a password for a user with a hashing algo
func (hp HashedPasswords) SetPassword(name, password string, hashAlgorithm HashAlgorithm) (err error) {
	if len(password) == 0 {
		return errors.New("passwords must not be empty, if you want to delete a user call RemoveUser")
	}
	hash := ""
	prefix := ""
	switch hashAlgorithm {
	case HashBCrypt:
		hash, err = hashBcrypt(password)
	case HashSHA:
		prefix = "{SHA}"
		hash = hashSha(password)
	}
	if err != nil {
		return err
	}
	hp[name] = prefix + hash
	return nil
}

// ParseHtpasswdFile load a htpasswd file
func ParseHtpasswdFile(file string) (passwords HashedPasswords, err error) {
	htpasswdBytes, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}
	if len(htpasswdBytes) > MaxHtpasswdFilesize {
		err = errors.New("this file is too large, use a database instead")
		return
	}
	return ParseHtpasswd(htpasswdBytes)
}

// ParseHtpasswd parse htpasswd bytes
func ParseHtpasswd(htpasswdBytes []byte) (passwords HashedPasswords, err error) {
	lines := strings.Split(string(htpasswdBytes), LineSeparator)
	passwords = make(map[string]string)
	for lineNumber, line := range lines {
		// scan lines
		line = strings.Trim(line, " ")
		if len(line) == 0 {
			// skipping empty lines
			continue
		}
		parts := strings.Split(line, PasswordSeparator)
		if len(parts) != 2 {
			err = errors.New(fmt.Sprintln("invalid line", lineNumber+1, "unexpected number of parts split by", PasswordSeparator, len(parts), "instead of 2 in\"", line, "\""))
			return
		}
		for i, part := range parts {
			parts[i] = strings.Trim(part, " ")
		}
		_, alreadyExists := passwords[parts[0]]
		if alreadyExists {
			err = errors.New("invalid htpasswords file - user " + parts[0] + " was already defined")
			return
		}
		passwords[parts[0]] = parts[1]
	}
	return
}

// SetHtpasswdHash set password hash for a user
func SetHtpasswdHash(file, name, hash string) error {
	passwords, err := ParseHtpasswdFile(file)
	if err != nil {
		return err
	}
	passwords[name] = hash
	return passwords.WriteToFile(file)
}

// RemoveUser remove an existing user from a file, returns an error, if the user does not \
// exist in the file
func RemoveUser(file, user string) error {
	passwords, err := ParseHtpasswdFile(file)
	if err != nil {
		return err
	}
	_, ok := passwords[user]
	if !ok {
		return errors.New("user did not exist in file")
	}
	delete(passwords, user)
	return passwords.WriteToFile(file)
}

// SetPasswordHash directly set a hash for a user in a file
func SetPasswordHash(file, user, hash string) error {
	if len(hash) == 0 {
		return errors.New("you might want to rethink your hashing algorithm, it left you with an empty hash")
	}
	passwords, err := ParseHtpasswdFile(file)
	if err != nil {
		return err
	}
	passwords[user] = hash
	return passwords.WriteToFile(file)
}

// SetPassword set password for a user with a given hashing algorithm
func SetPassword(file, name, password string, hashAlgorithm HashAlgorithm) error {
	_, err := os.Stat(file)
	passwords := HashedPasswords(map[string]string{})
	if err == nil {
		passwords, err = ParseHtpasswdFile(file)
		if err != nil {
			return err
		}
	}
	err = passwords.SetPassword(name, password, hashAlgorithm)
	if err != nil {
		return err
	}
	return passwords.WriteToFile(file)
}
