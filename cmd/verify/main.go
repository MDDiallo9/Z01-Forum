package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
)

const baseURL = "https://localhost:8000"

func main() {
	jar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Jar:       jar,
		Transport: tr,
	}

	// 1. Register
	fmt.Println("1. Registering new user...")
	vals := url.Values{}
	vals.Set("username", "autotest")
	vals.Set("email", "autotest@example.com")
	vals.Set("password", "password123")
	vals.Set("confirm_password", "password123")

	// We need to handle multipart upload for avatar even if empty
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("username", "autotest")
	writer.WriteField("email", "autotest@example.com")
	writer.WriteField("password", "password123")
	writer.WriteField("confirm_password", "password123")
	writer.Close()

	resp, err := client.Post(baseURL+"/register", writer.FormDataContentType(), body)
	if err != nil {
		fmt.Printf("Registration failed: %v\n", err)
		os.Exit(1)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusSeeOther {
		fmt.Printf("Registration returned status: %d\n", resp.StatusCode)
		// It might fail if user exists, so let's try login
	} else {
		fmt.Println("Registration successful (or redirected).")
	}

	// 2. Login
	fmt.Println("2. Logging in...")
	vals = url.Values{}
	vals.Set("username", "autotest")
	vals.Set("password", "password123")
	resp, err = client.PostForm(baseURL+"/login", vals)
	if err != nil {
		fmt.Printf("Login failed: %v\n", err)
		os.Exit(1)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusSeeOther {
		fmt.Printf("Login returned status: %d\n", resp.StatusCode)
		os.Exit(1)
	}
	fmt.Println("Login successful.")

	// 3. Create Post
	fmt.Println("3. Creating post...")
	body = &bytes.Buffer{}
	writer = multipart.NewWriter(body)
	writer.WriteField("title", "Automated Test Post")
	writer.WriteField("content", "This is a post created by the verification script.")
	writer.WriteField("categories", "1") // Assuming category ID 1 exists
	writer.Close()

	req, _ := http.NewRequest("POST", baseURL+"/post/create", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err = client.Do(req)
	if err != nil {
		fmt.Printf("Create post failed: %v\n", err)
		os.Exit(1)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusSeeOther {
		fmt.Printf("Create post returned status: %d\n", resp.StatusCode)
		os.Exit(1)
	}
	fmt.Println("Post created.")

	// 4. Verify Home Page
	fmt.Println("4. Verifying home page content...")
	resp, err = client.Get(baseURL + "/")
	if err != nil {
		fmt.Printf("Get home failed: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	content, _ := io.ReadAll(resp.Body)
	if strings.Contains(string(content), "Automated Test Post") {
		fmt.Println("SUCCESS: Found test post on home page!")
	} else {
		fmt.Println("FAILURE: Test post not found on home page.")
		os.Exit(1)
	}
}
