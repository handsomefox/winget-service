package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"golang.org/x/sys/windows"
)

// NOTE: Example commands
// winget export -o .\winget-export.json
// winget list --source winget > out.txt

const (
	exportFilename = "winget-export.json"
	listFilename   = "winget-list.txt"
	logFilename    = "log.txt"
)

// NOTE: Stolen from https://stackoverflow.com/a/59147866/12041778
func elevate() {
	verb := "runas"
	exe, _ := os.Executable()
	cwd, _ := os.Getwd()
	args := strings.Join(os.Args[1:], " ")

	verbPtr, _ := syscall.UTF16PtrFromString(verb)
	exePtr, _ := syscall.UTF16PtrFromString(exe)
	cwdPtr, _ := syscall.UTF16PtrFromString(cwd)
	argPtr, _ := syscall.UTF16PtrFromString(args)

	var showCmd int32 = 1 // SW_NORMAL

	err := windows.ShellExecute(0, verbPtr, exePtr, argPtr, cwdPtr, showCmd)
	if err != nil {
		fmt.Println(err)
	}
}

// NOTE: Stolen from https://stackoverflow.com/a/59147866/12041778
func isAdmin() bool {
	_, err := os.Open("\\\\.\\PHYSICALDRIVE0")
	if err != nil {
		fmt.Println("admin no")
		return false
	}
	fmt.Println("admin yes")
	return true
}

func main() {
	// if !isAdmin() {
	// 	elevate()
	// 	return
	// }

	start := time.Now()

	// %USERPROFILE%
	userprofile, err := os.UserHomeDir()
	if err != nil {
		return
	}

	// NOTE: Path to winget.exe
	wingetPath := path.Join(userprofile, "AppData", "Local", "Microsoft", "WindowsApps")

	// %USERPROFILE%/OneDrive
	home := path.Join(userprofile, "OneDrive")

	if err := os.Chdir(home); err != nil {
		log.Println("Couldn't move to home directory", err)
		return
	}

	wd, _ := os.Getwd()
	log.Printf("Current path: %s", wd)

	lfile, err := os.OpenFile(logFilename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o666)
	if err != nil {
		log.Println("Couldn't open file for logging", err)
		return
	}
	defer lfile.Close()

	log.SetOutput(lfile)

	wingetExec := wingetPath + "\\winget.exe"
	wingetExec = filepath.ToSlash(wingetExec)

	log.Printf("winget path: %v", wingetExec)

	log.Println("Starting the export")

	log.Println("Checking if export file already exists")
	if _, err := os.Stat(exportFilename); !errors.Is(err, os.ErrNotExist) {
		log.Println("Export file exists, removing...")
		os.Remove(exportFilename)
	} else {
		log.Println("Previous file was not found")
	}

	log.Println("Checking if list file already exists")
	if _, err := os.Stat(listFilename); !errors.Is(err, os.ErrNotExist) {
		log.Println("List file exists, removing...")
		os.Remove(listFilename)
	} else {
		log.Println("Previous file was not found")
	}

	log.Println("Executing export")

	exportCmd := exec.Command("powershell.exe", wingetExec, "export", "-o", path.Join(home, exportFilename))
	// NOTE: This avoid launching a window when invoking powershell
	exportCmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x08000000}

	log.Printf("Invoking %v", exportCmd.String())
	// Write output to log
	exportCmd.Stdout = lfile
	exportCmd.Stderr = lfile

	if err := exportCmd.Run(); err != nil {
		log.Print("Error while exporting: " + err.Error())
	}

	log.Println("Executing list")

	listCmd := exec.Command("powershell.exe", wingetExec, "list", "--source", "winget", ">", path.Join(home, listFilename))
	// NOTE: This avoid launching a window when invoking powershell
	listCmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x08000000}

	log.Printf("Invoking %v", listCmd.String())
	// Write output to log
	listCmd.Stdout = lfile
	listCmd.Stderr = lfile

	if err := listCmd.Run(); err != nil {
		log.Print("Error while listing: " + err.Error())
	}

	log.Printf("Finished backup in: %vms", time.Since(start).Milliseconds())
}
