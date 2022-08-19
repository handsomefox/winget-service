# winget-service

## Why

You can run this application on schedule / startup (shell:startup folder) to backup your application list to OneDrive (or another folder).

## NOTE

1. It may require admin rights, uncomment admin code if needed

2. To hide console window, compile using

```bash
go build --ldflags "-H=windowsgui -s -w" main.go
```
