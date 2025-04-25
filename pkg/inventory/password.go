package inventory

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"golang.org/x/term"
	"gopkg.in/ini.v1"
	"os"
	"strings"
)

const ezMonitorEncDelimiter = "$EZ_MONITOR_ENCRYPTED;"

func BeginPasswordEncryptFlow(hostToAddEncryptedPassword, filename string) error {
	cfg, err := ini.Load(filename)
	if err != nil {
		return fmt.Errorf("failed to load ini data: %s", err)
	}
	hostSection, err := cfg.GetSection(hostToAddEncryptedPassword)
	if err != nil { // This func only returns an error if the section does not exist
		// If the section does not exist, we will add it
		fmt.Printf("There is currently no entry in your inventory file for the host %s\n", hostToAddEncryptedPassword)
		fmt.Println("Would you like an entry to be added for this host? [y/n]")
		err = exitIfNoResponse()
		if err != nil {
			return err
		}
		hostSection, err = cfg.NewSection(hostToAddEncryptedPassword)
		if err != nil {
			return fmt.Errorf("failed to create new entry in inventory file for new host %s: %s", hostToAddEncryptedPassword, err)
		}
	}

	if hostSection.HasKey("password") {
		fmt.Printf("There is already a password entry in your inventory file for the host %s\n", hostToAddEncryptedPassword)
		fmt.Println("Would you like to replace the password entry for this host? [y/n]")
		err = exitIfNoResponse()
		if err != nil {
			return err
		}
	}

	fmt.Printf("Enter the password you would like to set for the host %s:\n", hostToAddEncryptedPassword)
	hostPasswordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println() // Print a newline after password input
	if err != nil {
		return fmt.Errorf("failed to read host password: %s", err)
	}
	hostPassword := string(hostPasswordBytes)

	fmt.Println("Enter your encryption password. This encryption password should be used for all passwords in this file. " +
		"Make sure that you remember it as it will be required to decrypt the passwords when running ez-monitor in the future.")
	encPasswordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println() // Print a newline after password input
	if err != nil {
		return fmt.Errorf("failed to read encryption password: %s", err)
	}
	encPassword := string(encPasswordBytes)

	encryptedHostPassword, err := encrypt(hostPassword, encPassword)
	if err != nil {
		return fmt.Errorf("failed to encrypt password: %s", err)
	}
	hostSection.Key("password").SetValue(encryptedHostPassword)

	err = addOrReplacePasswordValue(filename, hostToAddEncryptedPassword, encryptedHostPassword)
	if err != nil {
		return fmt.Errorf("failed to save ini data: %s", err)
	}

	return nil
}

func encrypt(data, passphrase string) (string, error) {
	key := sha256.Sum256([]byte(passphrase)) // Derive a 256-bit key from the passphrase
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}

	// Use AES-GCM for encryption
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize()) // Generate a random nonce
	ciphertext := aesGCM.Seal(nonce, nonce, []byte(data), nil)
	return fmt.Sprintf("%s%s", ezMonitorEncDelimiter, hex.EncodeToString(ciphertext)), nil
}

func decrypt(encryptedData, passphrase string) (string, error) {
	parsedEncryptedData, found := strings.CutPrefix(encryptedData, ezMonitorEncDelimiter)
	if !found {
		return "", fmt.Errorf("failed to parse encrypted data")
	}

	key := sha256.Sum256([]byte(passphrase))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	data, err := hex.DecodeString(parsedEncryptedData)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(plaintext)), nil
}

func exitIfNoResponse() error {
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		return fmt.Errorf("failed to read response: %s\n", err)
	}
	if strings.ToLower(response) != "y" {
		os.Exit(0)
	}
	return nil
}

// addOrReplacePasswordValue is a customer implementation seperate from the ini package.
// This is necessary to preserve the files formatting and comments as the ini package will reformat
// The file if used to save a new config when using config.SaveTo()
// This function which will perform the following actions
// 1. Load in the ini inventory file
// 2. Add or replace the password for the appropriate host. If the host does not exist, it will add a host value at the bottom of the file
// 3. Replace the file with the new contents
func addOrReplacePasswordValue(filename, hostToAddEncryptedPassword, encryptedHostPassword string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %s", err)
	}

	lines := strings.Split(string(content), "\n")
	var updatedLines []string
	sectionFound := false
	passwordUpdated := false
	sectionHeader := "[" + hostToAddEncryptedPassword + "]"
	newPasswordLine := fmt.Sprintf("password = `%s`", encryptedHostPassword)

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmedLine := strings.TrimSpace(line)

		// Check if the current line is the section header
		if trimmedLine == sectionHeader {
			sectionFound = true
			updatedLines = append(updatedLines, line)

			// Process the section to add or replace the password
			for i++; i < len(lines); i++ {
				trimmedLine = strings.TrimSpace(lines[i])

				// Stop processing if a new section starts
				if strings.HasPrefix(trimmedLine, "[") && strings.HasSuffix(trimmedLine, "]") {
					updatedLines = append(updatedLines, newPasswordLine)
					passwordUpdated = true
					break
				}

				// Replace the password if it exists
				if strings.HasPrefix(trimmedLine, "password") {
					updatedLines = append(updatedLines, newPasswordLine)
					passwordUpdated = true
				} else {
					updatedLines = append(updatedLines, lines[i])
				}
			}
		}

		// Add the current line to the updated lines if not processed
		if !passwordUpdated {
			updatedLines = append(updatedLines, line)
		}
	}

	// If the section was not found, add it at the end
	if !sectionFound {
		updatedLines = append(updatedLines, sectionHeader)
		updatedLines = append(updatedLines, newPasswordLine)
	}

	// Write the updated content back to the file
	stat, err := os.Stat(filename)
	if err != nil {
		return fmt.Errorf("failed to stat file: %s", err)
	}

	err = os.WriteFile(filename, []byte(strings.Join(updatedLines, "\n")), stat.Mode())
	if err != nil {
		return fmt.Errorf("failed to write updated file: %s", err)
	}

	return nil
}
