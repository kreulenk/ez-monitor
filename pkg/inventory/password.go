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

	err = cfg.SaveTo(filename)
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
