package main

import (
	"context"
	"fmt"
	"github.com/google/go-github/v65/github"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

const (
	DAPR_GITHUB_ORG_ID          = 51932459
	DAPR_GITHUB_RELEASE_TEAM_ID = 4237823
)

type DaprCore struct {
	Name  string
	Value string
}

type DaprSDK struct {
	Name  string
	Value string
}

func ParseMarkdown(markdown string) ([]DaprCore, []DaprSDK, error) {
	var daprCore []DaprCore
	var sdks []DaprSDK

	lines := strings.Split(markdown, "\n")

	// Regex to parse DaprCore lines
	coreRegex := regexp.MustCompile(`^\* ([^:]+):\s*(.*)$`)
	// Regex to parse Dapr SDK lines
	sdkRegex := regexp.MustCompile(`^\s*\* ([^:]+):\s*(.*)$`)

	inSDKs := false
	for _, line := range lines {
		if strings.HasPrefix(line, "* SDKs:") {
			inSDKs = true
			continue
		}

		if inSDKs {
			matches := sdkRegex.FindStringSubmatch(line)
			if len(matches) == 3 {
				sdks = append(sdks, DaprSDK{Name: matches[1], Value: strings.TrimSpace(matches[2])})
			}
		} else {
			matches := coreRegex.FindStringSubmatch(line)
			if len(matches) == 3 {
				daprCore = append(daprCore, DaprCore{Name: matches[1], Value: strings.TrimSpace(matches[2])})
			}
		}
	}

	if !inSDKs {
		return nil, nil, fmt.Errorf("no SDKs section found")
	}

	return daprCore, sdks, nil
}

func CheckAllFields(core []DaprCore, sdks []DaprSDK) error {
	// Define all required Dapr core values here
	requiredDaprCore := map[string]bool{
		"RC":                      false,
		"dapr/components-contrib": false,
		"dapr/dapr":               false,
		"dapr/cli":                false,
		"dapr/dashboard":          false,
	}

	// Define all required SDK values here
	requiredDaprSDKs := map[string]bool{
		"go":     false,
		"rust":   false,
		"python": false,
		"dotnet": false,
		"java":   false,
		"js":     false,
	}

	for _, c := range core {
		if _, ok := requiredDaprCore[c.Name]; ok {
			requiredDaprCore[c.Name] = true
		}
	}

	for _, s := range sdks {
		if _, ok := requiredDaprSDKs[s.Name]; ok {
			requiredDaprSDKs[s.Name] = true
		}
	}

	for c, present := range requiredDaprCore {
		if !present {
			return fmt.Errorf("missing core definition: %s", c)
		}
	}

	for s, present := range requiredDaprSDKs {
		if !present {
			return fmt.Errorf("missing SDK: %s", s)
		}
	}

	return nil
}

func ValidateInput(daprCore []DaprCore, sdks []DaprSDK) error {
	for _, c := range daprCore {
		if !isValidInput(c.Name) || !isValidInput(c.Value) {
			return fmt.Errorf("invalid input for dapr core: %s", c.Name)
		}
	}
	for _, s := range sdks {
		if !isValidInput(s.Name) || !isValidInput(s.Value) {
			return fmt.Errorf("invalid input for SDK: %s", s.Name)
		}
	}
	return nil
}

func isValidInput(input string) bool {
	matched, _ := regexp.MatchString(`^[A-Za-z0-9./-]+$`, input)
	return matched
}

func main() {
	if !isIssueEvent() {
		log.Println("Not an issues event, exiting")
		os.Exit(0)
	}

	client := github.NewClient(nil).WithAuthToken(os.Getenv("GITHUB_API_TOKEN"))

	triggeringEvent, err := getGithubIssuesEventFromEnv(nil)
	if err != nil {
		log.Fatalf("failed to get event: %v", err)
	}

	if triggeringEvent.GetSender() == nil {
		log.Fatalln("no sender(actor) found in event")
	}

	ctx := context.Background()
	// TODO: revert to the constants
	triggeringActorMembership, resp, err := client.Teams.GetTeamMembershipByID(ctx, 182497202, 11049175, triggeringEvent.GetSender().GetLogin())
	log.Printf("getting membership for: %v", triggeringEvent.GetSender().GetLogin())
	if err != nil {
		log.Fatalf("failed to retrieve membership: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("user is not a member: %s", triggeringEvent.GetSender().GetLogin())
		os.Exit(0) // successful run - but exit
	}

	// confirm active membership
	if triggeringActorMembership.GetState() != "active" {
		log.Printf("user membership is not active: %s", triggeringEvent.GetSender().GetLogin())
		os.Exit(0) // successful run - but exit
	}

	markdown := triggeringEvent.GetIssue().GetBody()

	core, sdks, err := ParseMarkdown(markdown)
	if err != nil {
		log.Fatalf("failed to parse markdown: %v", err)
	}

	if err := CheckAllFields(core, sdks); err != nil {
		log.Printf("checking of all fields failed: %v\n", err)
	}

	if err := ValidateInput(core, sdks); err != nil {
		log.Printf("failed to validate: %v\n", err)
	}

	log.Println("Core:")
	for _, c := range core {
		log.Printf("- %s: %s\n", c.Name, c.Value)
	}

	log.Println("\nSDKs:")
	for _, sdk := range sdks {
		log.Printf("- %s: %s\n", sdk.Name, sdk.Value)
	}

	// get commit status
	status, resp, err := client.Repositories.GetCombinedStatus(ctx, "dapr-test", "dapr", "main", nil)
	if err != nil {
		log.Fatalf("dapr-test/dapr not found: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("dapr-test/dapr commit status request failed - http status: %v", resp.Status)
	}

	log.Printf("status: %v, totalCount: %v, sha: %v", status.GetState(), status.GetTotalCount(), status.GetSHA())
}
