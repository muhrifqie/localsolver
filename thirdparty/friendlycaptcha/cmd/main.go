package main

import (
	"fmt"
	"os"

	"github.com/xKiian/friendlycaptcha/api"
)

const (
	defaultSitekey = "FCMU08117KKFLBRL"
	defaultAPIKey  = ""
)

func main() {
	sitekey := defaultSitekey
	if len(os.Args) > 1 {
		sitekey = os.Args[1]
	}

	proxy := ""
	if len(os.Args) > 2 {
		proxy = os.Args[2]
	}

	client, err := api.NewClient(sitekey, proxy)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	token, err := client.Solve()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(token)

	if defaultAPIKey != "" {
		fmt.Println("[*] Verifying token...")
		result, err := client.Verify(token, defaultAPIKey)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Verify error: %v\n", err)
			os.Exit(1)
		}

		if result.Success {
			fmt.Println("[+] Token is VALID")
		} else {
			fmt.Println("[-] Token is INVALID")
			if result.Error != nil {
				fmt.Printf("    %s: %s\n", result.Error.ErrorCode, result.Error.Detail)
			}
		}
	}
}
