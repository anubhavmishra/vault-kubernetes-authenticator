package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/pkg/errors"
)

var (
	vaultAddr         string
	vaultK8SMountPath string
)

func main() {
	vaultAddr = os.Getenv("VAULT_ADDR")
	if vaultAddr == "" {
		vaultAddr = "https://127.0.0.1:8200"
	}

	vaultK8SMountPath = os.Getenv("VAULT_K8S_MOUNT_PATH")
	if vaultK8SMountPath == "" {
		vaultK8SMountPath = "kubernetes"
	}

	role := os.Getenv("VAULT_ROLE")
	if role == "" {
		log.Fatal("missing VAULT_ROLE")
	}

	dest := os.Getenv("TOKEN_DEST_PATH")
	if dest == "" {
		dest = "/.vault-token"
	}

	saPath := os.Getenv("SERVICE_ACCOUNT_PATH")
	if saPath == "" {
		saPath = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	}

	// Read the JWT token from disk
	jwt, err := readJwtToken(saPath)
	if err != nil {
		log.Fatal(err)
	}

	// Authenticate to vault using the jwt token
	token, err := authenticate(role, jwt)
	if err != nil {
		log.Fatal(err)
	}

	// Persist the vault token to disk
	if err := saveToken(token, dest); err != nil {
		log.Fatal(err)
	}

	log.Printf("successfully stored vault token at %s", dest)

	os.Exit(0)
}

func readJwtToken(path string) (string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", errors.Wrap(err, "failed to read jwt token")
	}

	return string(bytes.TrimSpace(data)), nil
}

func authenticate(role, jwt string) (string, error) {

	client := &http.Client{}

	addr := vaultAddr + "/v1/auth/" + vaultK8SMountPath + "/login"
	body := fmt.Sprintf(`{"role": "%s", "jwt": "%s"}`, role, jwt)

	req, err := http.NewRequest(http.MethodPost, addr, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		return "", errors.Wrap(err, "failed to create request")
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "failed to login")
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		var b bytes.Buffer
		io.Copy(&b, resp.Body)
		return "", fmt.Errorf("failed to get successful response: %#v, %s",
			resp, b.String())
	}

	var s struct {
		Auth struct {
			ClientToken string `json:"client_token"`
		} `json:"auth"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&s); err != nil {
		return "", errors.Wrap(err, "failed to read body")
	}

	return s.Auth.ClientToken, nil
}

func saveToken(token, dest string) error {
	if err := ioutil.WriteFile(dest, []byte(token), 0600); err != nil {
		return errors.Wrap(err, "failed to save token")
	}
	return nil
}
