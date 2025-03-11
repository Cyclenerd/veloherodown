package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	myUserAgent       = "veloherodown/2.0"
	downloadRetries   = 3
	downloadBreaktime = 1             // seconds
	configFileName    = ".veloherorc" // In current directory
)

type Config struct {
	SsoKey string
}

func main() {
	// Check for help flags
	if len(os.Args) > 1 {
		arg := os.Args[1]
		if arg == "--help" || arg == "-h" {
			showHelp()
			return
		}
	}

	// Check configuration file
	config, err := loadConfig()
	if err != nil {
		if strings.Contains(err.Error(), "configuration file not found") {
			// Configuration file not found, prompt user for SSO key
			config.SsoKey = promptForSsoKey()

			// Save the SSO key to the configuration file
			if err := saveSsoKey(config.SsoKey); err != nil {
				exitWithFailure(fmt.Sprintf("Failed to save SSO key: %v", err))
			}
		} else {
			exitWithFailure(err.Error())
		}
	}

	// Test if SSO key is available
	if config.SsoKey == "" {
		echoSsoKeyHelp()
	}

	// Check login
	if !checkSsoLogin(config.SsoKey) {
		exitWithFailure("Login failed! Single Sign-on key not found or expired. Please get a new one.")
	}

	// Check if current directory is writable
	testFile := ".write_test"
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		exitWithFailure("Can not write to current directory. Please correct this.")
	}
	os.Remove(testFile)

	// Get formats to export from command line
	formats := parseFormats(os.Args[1:])
	if len(formats) == 0 {
		formats = []string{"pwx"}
	}

	formatStr := strings.Join(formats, ", ")
	if len(formats) > 1 {
		fmt.Printf("Will export data to formats '%s'...\n", formatStr)
	} else {
		fmt.Printf("Will export data to format '%s'...\n", formatStr)
	}

	// File with the specification of the last export
	lastExportFile := ".velohero_last_export.do_not_remove"

	// Get timestamp from last export
	lastTimestamp := int64(0)
	if _, err := os.Stat(lastExportFile); err == nil {
		content, err := os.ReadFile(lastExportFile)
		if err == nil {
			for _, line := range strings.Split(string(content), "\n") {
				if strings.HasPrefix(line, "VELOHERO_LAST_TIMESTAMP=") {
					timestampStr := strings.TrimPrefix(line, "VELOHERO_LAST_TIMESTAMP=")
					lastTimestamp, _ = strconv.ParseInt(timestampStr, 10, 64)
				}
			}
		}
	} else {
		// Create file if it doesn't exist
		if err := os.WriteFile(lastExportFile, []byte("VELOHERO_LAST_TIMESTAMP=0\n"), 0644); err != nil {
			exitWithFailure(fmt.Sprintf("Can not write file '%s'. The last download will be noted in this file.", lastExportFile))
		}
	}

	// Get list
	exportListFile := ".velohero_export_workouts.csv"
	lastChangeTime := time.Unix(lastTimestamp, 0).Format("2006-01-02 15:04:05")
	fmt.Printf("Get list of workouts since %s. Please wait...\n", lastChangeTime)

	if err := downloadWorkoutList(config.SsoKey, lastTimestamp, exportListFile); err != nil {
		exitWithFailure("Can not download list of files to be exported.")
	}
	defer os.Remove(exportListFile)

	// Count workouts
	workouts, err := parseWorkoutList(exportListFile)
	if err != nil {
		exitWithFailure(fmt.Sprintf("Error parsing workout list: %v", err))
	}

	if len(workouts) == 0 {
		fmt.Println("\nDone. No new files.")
		return
	}

	// Download workout files
	for i, workout := range workouts {
		fmt.Printf("Downloading file %d of %d with ID %s (%s %s). Please wait...\n",
			i+1, len(workouts), workout.ID, workout.Date, workout.StartTime)

		// Export for each format
		for _, format := range formats {
			exportActivity(config.SsoKey, workout.ID, format)
		}
		fmt.Println()
	}

	// Save previous and last export timestamp
	currentTimestamp := time.Now().Unix()
	content := fmt.Sprintf("VELOHERO_PREV_TIMESTAMP=%d\nVELOHERO_LAST_TIMESTAMP=%d\n", lastTimestamp, currentTimestamp)
	if err := os.WriteFile(lastExportFile, []byte(content), 0644); err != nil {
		exitWithFailure(fmt.Sprintf("Could not write to file '%s'", lastExportFile))
	}

	fmt.Println("\nDone. All downloaded.")
}

type Workout struct {
	ID        string
	Date      string
	StartTime string
	Duration  string
}

// promptForSsoKey asks the user to input their SSO key
func promptForSsoKey() string {
	fmt.Println("No configuration file found.")
	fmt.Println("Please go to https://app.velohero.com/sso to get your single sign-on key.")
	fmt.Print("Enter your Velo Hero SSO key: ")

	reader := bufio.NewReader(os.Stdin)
	ssoKey, _ := reader.ReadString('\n')

	// Trim whitespace and newlines
	ssoKey = strings.TrimSpace(ssoKey)

	return ssoKey
}

// saveSsoKey saves the SSO key to the configuration file
func saveSsoKey(ssoKey string) error {
	content := fmt.Sprintf("VELOHERO_SSO_KEY=%s\n", ssoKey)
	return os.WriteFile(configFileName, []byte(content), 0644)
}

func loadConfig() (Config, error) {
	// Use current directory for config file
	configPath := configFileName

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return Config{}, fmt.Errorf("configuration file not found")
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, fmt.Errorf("could not read configuration file: %v", err)
	}

	var config Config
	for _, line := range strings.Split(string(content), "\n") {
		if strings.HasPrefix(line, "VELOHERO_SSO_KEY=") {
			config.SsoKey = strings.TrimPrefix(line, "VELOHERO_SSO_KEY=")
		}
	}

	return config, nil
}

func parseFormats(args []string) []string {
	validFormats := map[string]bool{
		"json": true,
		"pwx":  true,
		"csv":  true,
		"gpx":  true,
		"kml":  true,
		"tcx":  true,
	}

	var formats []string
	for _, arg := range args {
		if validFormats[arg] {
			// Check if format is already in the list
			found := false
			for _, f := range formats {
				if f == arg {
					found = true
					break
				}
			}
			if !found {
				formats = append(formats, arg)
			}
		} else {
			exitWithFailure(fmt.Sprintf("Unknown format '%s'. Known formats are 'json', 'pwx', 'csv', 'gpx', 'kml' and 'tcx'.", arg))
		}
	}
	return formats
}

func parseWorkoutList(filename string) ([]Workout, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var workouts []Workout
	scanner := bufio.NewScanner(file)
	re := regexp.MustCompile(`^\d+$`)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ";")
		if len(parts) >= 4 && re.MatchString(parts[0]) {
			workouts = append(workouts, Workout{
				ID:        parts[0],
				Date:      parts[1],
				StartTime: parts[2],
				Duration:  parts[3],
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return workouts, nil
}

func checkSsoLogin(ssoKey string) bool {
	url := "https://app.velohero.com/sso"

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.WriteField("sso", ssoKey)
	_ = writer.WriteField("view", "json")
	writer.Close()

	req, err := http.NewRequest("POST", url, &body)
	if err != nil {
		return false
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("User-Agent", myUserAgent)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

func downloadWorkoutList(ssoKey string, lastTimestamp int64, outputFile string) error {
	url := "https://app.velohero.com/export/workouts/csv"

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.WriteField("sso", ssoKey)
	_ = writer.WriteField("last_change_epoch", strconv.FormatInt(lastTimestamp, 10))
	writer.Close()

	req, err := http.NewRequest("POST", url, &body)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("User-Agent", myUserAgent)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	out, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	// Add newline at the end
	_, err = out.WriteString("\n")
	return err
}

func exportActivity(ssoKey, workoutID, format string) {
	exportFile := workoutID + "." + format

	// Skip if file already exists
	if _, err := os.Stat(exportFile); err == nil {
		return
	}

	url := fmt.Sprintf("https://app.velohero.com/export/activity/%s/%s", format, workoutID)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.WriteField("sso", ssoKey)
	writer.Close()

	req, err := http.NewRequest("POST", url, &body)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("User-Agent", myUserAgent)

	client := &http.Client{}

	// Implement retries
	var resp *http.Response
	for i := 0; i < downloadRetries; i++ {
		resp, err = client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			break
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(time.Duration(downloadBreaktime) * time.Second)
	}

	if err != nil || resp.StatusCode != http.StatusOK {
		fmt.Printf("Failed to download workout %s in format %s\n", workoutID, format)
		if resp != nil {
			resp.Body.Close()
		}
		return
	}
	defer resp.Body.Close()

	out, err := os.Create(exportFile)
	if err != nil {
		fmt.Printf("Error creating file: %v\n", err)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		fmt.Printf("Error writing to file: %v\n", err)
		return
	}

	// Give server a break
	time.Sleep(time.Duration(downloadBreaktime) * time.Second)
}

// showHelp displays the usage information for the program
func showHelp() {
	fmt.Printf(`
veloherodown - Download your Velo Hero data

Usage: %s [OPTIONS] [FORMAT...]

OPTIONS:
  --help, -h,    Show this help message and exit

FORMATS:
  json           Velo Hero generic JSON format (with all details)
  pwx            Training Peaks PWX file with laps (can be processed by Golden Cheetah)
  csv            Comma-Seperated Values CSV file
  gpx            GPX file (only the geo coordinates)
  kml            Google Earth KML file
  tcx            Garmin TCX file

If no format is specified, 'pwx' is used by default.
Multiple formats can be specified to download in several formats.

Examples:
  %s             Download in PWX format
  %s json        Download in JSON format
  %s json pwx    Download in both JSON and PWX formats

The first time all files are downloaded.
For further calls only changes and new files are downloaded.
All files are saved in the current working directory.

Configuration:
  The program looks for a '.veloherorc' file in the current directory.
  If not found, it will prompt you to enter your SSO key.
  You can get your SSO key at https://app.velohero.com/sso

`, os.Args[0], os.Args[0], os.Args[0], os.Args[0])
}

func echoSsoKeyHelp() {
	fmt.Println()
	fmt.Printf("To use the '%s' script, please go to https://app.velohero.com/sso\n", os.Args[0])
	fmt.Println("and get yourself a private single sign-on key. That's the long string.")
	fmt.Println()
	fmt.Println("Then create a file '.veloherorc' in the current directory containing")
	fmt.Println()
	fmt.Println("----- snip -------------------------------------------------------------")
	fmt.Println()
	fmt.Println("VELOHERO_SSO_KEY=insert your own")
	fmt.Println()
	fmt.Println("----- snap -------------------------------------------------------------")
	fmt.Println()
	fmt.Println("Important: Do not use spaces!")
	fmt.Println()
	os.Exit(1)
}

func exitWithFailure(message string) {
	fmt.Println()
	fmt.Printf("FAILURE: %s\n", message)
	fmt.Println()
	os.Exit(9)
}
