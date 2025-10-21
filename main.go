package main

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/blang/semver"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/corpix/uarand"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
	"gopkg.in/yaml.v3"
)

const version = "1.0.0"

//go:embed services.yaml
var servicesYAML embed.FS

type ServiceConfig struct {
	Name           string            `yaml:"name"`
	Method         string            `yaml:"method"`
	URL            string            `yaml:"url"`
	Headers        map[string]string `yaml:"headers"`
	AuthType       string            `yaml:"auth_type"`
	AuthUser       string            `yaml:"auth_user"`
	AuthPass       string            `yaml:"auth_pass"`
	SuccessStatus  int               `yaml:"success_status"`
	ResponseType   string            `yaml:"response_type"`
	ResponseFields []string          `yaml:"response_fields"`
	DetailsFormat  string            `yaml:"details_format"`
	SuccessField   string            `yaml:"success_field"`
	ErrorField     string            `yaml:"error_field"`
	RequiresSecret bool              `yaml:"requires_secret"`
	SecretName     string            `yaml:"secret_name"`
	SDKType        string            `yaml:"sdk_type"`
	Service        string            `yaml:"service"`
	Operation      string            `yaml:"operation"`
	Message        string            `yaml:"message"`
	Details        string            `yaml:"details"`
}

type ServicesConfig struct {
	Services map[string]ServiceConfig `yaml:"services"`
}

type VerificationResult struct {
	Service   string `json:"service"`
	Key       string `json:"key,omitempty"`
	Valid     bool   `json:"valid"`
	Message   string `json:"message"`
	Details   string `json:"details,omitempty"`
	Timestamp string `json:"timestamp"`
}

var (
	servicesConfig ServicesConfig
	successStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	errorStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	dimStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	highlightStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
)

func init() {
	log.SetTimeFormat("15:04:05")
	log.SetLevel(log.WarnLevel)
	log.SetReportCaller(false)
	loadServicesConfig()
}

func loadServicesConfig() {
	data, err := servicesYAML.ReadFile("services.yaml")
	if err != nil {
		log.Fatal("Failed to read services.yaml", "error", err)
	}

	if err := yaml.Unmarshal(data, &servicesConfig); err != nil {
		log.Fatal("Failed to parse services.yaml", "error", err)
	}
}

func main() {
	service, key, secret, jsonOutput, listServices, showHelp, showVersion, doUpdate := parseFlags()
	if showHelp {
		displayHelp()
		return
	}
	if showVersion {
		displayVersion()
		return
	}
	if doUpdate {
		performUpdate()
		return
	}
	if listServices {
		displayServices()
		return
	}

	result := verifyAPIKey(service, key, secret)
	if jsonOutput {
		json.NewEncoder(os.Stdout).Encode(result)
	} else {
		displayResult(result)
	}
	if !result.Valid {
		os.Exit(1)
	}
}

func parseFlags() (string, string, string, bool, bool, bool, bool, bool) {
	service := flag.String("s", "", "service type")
	key := flag.String("k", "", "api key")
	secret := flag.String("secret", "", "secret key")
	jsonOutput := flag.Bool("json", false, "json output")
	listServices := flag.Bool("list", false, "list services")
	showHelp := flag.Bool("h", false, "help")
	showVersion := flag.Bool("version", false, "show version")
	doUpdate := flag.Bool("update", false, "update to latest version")
	flag.Parse()

	if *showHelp {
		return "", "", "", false, false, true, false, false
	}
	if *showVersion {
		return "", "", "", false, false, false, true, false
	}
	if *doUpdate {
		return "", "", "", false, false, false, false, true
	}
	if *listServices {
		return "", "", "", false, true, false, false, false
	}
	if *service == "" || *key == "" {
		displayHelp()
		os.Exit(0)
	}
	return *service, *key, *secret, *jsonOutput, false, false, false, false
}

func displayHelp() {
	cmdStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
	argStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	flagStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	requiredStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	
	fmt.Println()
	fmt.Println(successStyle.Render(" example:"))
	fmt.Printf("    %s -s %s -k %s\n", cmdStyle.Render("roq"), argStyle.Render("github"), argStyle.Render("ghp_xxxxxxxxxxxx"))
	fmt.Printf("    %s -s %s -json\n\n", cmdStyle.Render("roq"), argStyle.Render("trello"))
	
	fmt.Println(successStyle.Render(" options:"))
	fmt.Printf("    %s       service type %s\n", flagStyle.Render("-s"), requiredStyle.Render("(required)"))
	fmt.Printf("    %s       api key to verify %s\n", flagStyle.Render("-k"), requiredStyle.Render("(required)"))
	fmt.Printf("    %s  secret key %s\n", flagStyle.Render("-secret"), argStyle.Render("(required for aws)"))
	fmt.Printf("    %s    output in json format\n", flagStyle.Render("-json"))
	fmt.Printf("    %s    list all supported services\n", flagStyle.Render("-list"))
	fmt.Printf("    %s show version\n", flagStyle.Render("-version"))
	fmt.Printf("    %s  update to latest version\n", flagStyle.Render("-update"))
	fmt.Printf("    %s       show this help message\n\n", flagStyle.Render("-h"))
	
	fmt.Println(argStyle.Render("use responsibly and only on authorized targets!"))
	fmt.Println()
}

func displayVersion() {
	fmt.Println()
	fmt.Printf("%s %s\n", highlightStyle.Render("roq"), dimStyle.Render("v"+version))
	fmt.Println()
}

func performUpdate() {
	fmt.Println()
	fmt.Println(highlightStyle.Render("checking for updates..."))
	
	latest, found, err := selfupdate.DetectLatest("1hehaq/roq")
	if err != nil {
		fmt.Printf("%s %s\n", errorStyle.Render("✗"), dimStyle.Render("error checking for updates: "+err.Error()))
		fmt.Println()
		os.Exit(1)
	}
	
	if !found {
		fmt.Printf("%s %s\n", errorStyle.Render("✗"), dimStyle.Render("no releases found"))
		fmt.Println()
		os.Exit(1)
	}
	
	currentVersion := "v" + version
	v, err := semver.ParseTolerant(strings.TrimPrefix(currentVersion, "v"))
	if err != nil {
		fmt.Printf("%s %s\n", errorStyle.Render("✗"), dimStyle.Render("invalid version format: "+err.Error()))
		fmt.Println()
		os.Exit(1)
	}
	
	if !latest.Version.GT(v) {
		fmt.Printf("%s %s\n", successStyle.Render("✓"), dimStyle.Render("already up to date ("+currentVersion+")"))
		fmt.Println()
		return
	}
	
	exe, err := os.Executable()
	if err != nil {
		fmt.Printf("%s %s\n", errorStyle.Render("✗"), dimStyle.Render("could not locate executable: "+err.Error()))
		fmt.Println()
		os.Exit(1)
	}
	
	fmt.Printf("  %s → %s\n", dimStyle.Render(currentVersion), highlightStyle.Render(latest.Version.String()))
	fmt.Println()
	fmt.Print(dimStyle.Render("  updating... "))
	
	if err := selfupdate.UpdateTo(latest.AssetURL, exe); err != nil {
		fmt.Printf("%s\n", errorStyle.Render("failed"))
		fmt.Printf("  %s\n", dimStyle.Render("error: "+err.Error()))
		fmt.Println()
		os.Exit(1)
	}
	
	fmt.Printf("%s\n", successStyle.Render("done"))
	fmt.Println()
	fmt.Println(dimStyle.Render("  restart roq to use the new version"))
	fmt.Println()
}

func displayServices() {
	fmt.Println()
	fmt.Println(highlightStyle.Render("supported services:"))
	fmt.Println()
	for serviceName, serviceConfig := range servicesConfig.Services {
		secretInfo := ""
		if serviceConfig.RequiresSecret {
			secretInfo = dimStyle.Render(" (requires secret)")
		}
		fmt.Printf("  • %s - %s%s\n", serviceName, serviceConfig.Name, secretInfo)
	}
	fmt.Println()
}

func displayResult(result VerificationResult) {
	fmt.Println()
	if result.Valid {
		fmt.Printf("%s %s\n", successStyle.Render("✓"), strings.ToLower(result.Service))
		if result.Details != "" {
			fmt.Printf("  %s\n", dimStyle.Render(strings.ToLower(result.Details)))
		}
	} else {
		fmt.Printf("%s %s\n", errorStyle.Render("✗"), strings.ToLower(result.Service))
		fmt.Printf("  %s\n", dimStyle.Render(strings.ToLower(result.Message)))
	}
	fmt.Println()
}

func verifyAPIKey(service, key, secret string) VerificationResult {
	serviceConfig, exists := servicesConfig.Services[strings.ToLower(service)]
	if !exists {
		return VerificationResult{
			Service:   strings.ToLower(service),
			Valid:     false,
			Message:   fmt.Sprintf("unsupported service: %s", service),
			Timestamp: time.Now().Format(time.RFC3339),
		}
	}

	result := VerificationResult{
		Service:   strings.ToLower(serviceConfig.Name),
		Key:       maskKey(key),
		Timestamp: time.Now().Format(time.RFC3339),
	}

	switch serviceConfig.Method {
	case "GET", "POST":
		return verifyHTTP(serviceConfig, key, result)
	case "SDK":
		if serviceConfig.SDKType == "aws" {
			return verifyAWS(key, secret, result)
		}
	case "MANUAL":
		result.Valid = false
		result.Message = strings.ToLower(serviceConfig.Message)
		result.Details = strings.ToLower(serviceConfig.Details)
		return result
	}

	result.Valid = false
	result.Message = "verification method not implemented"
	return result
}

func verifyHTTP(serviceConfig ServiceConfig, key string, result VerificationResult) VerificationResult {
	url := renderTemplate(serviceConfig.URL, map[string]string{"Key": key})
	req, err := http.NewRequest(serviceConfig.Method, url, nil)
	if err != nil {
		result.Valid = false
		result.Message = "failed to create request"
		return result
	}

	for headerKey, headerValue := range serviceConfig.Headers {
		rendered := renderTemplate(headerValue, map[string]string{
			"Key":       key,
			"UserAgent": uarand.GetRandom(),
		})
		req.Header.Set(headerKey, rendered)
	}

	if serviceConfig.AuthType == "basic" {
		authUser := renderTemplate(serviceConfig.AuthUser, map[string]string{"Key": key})
		authPass := renderTemplate(serviceConfig.AuthPass, map[string]string{"Key": key})
		req.SetBasicAuth(authUser, authPass)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		result.Valid = false
		result.Message = "request failed: " + err.Error()
		return result
	}
	defer resp.Body.Close()

	if resp.StatusCode == serviceConfig.SuccessStatus {
		if serviceConfig.ResponseType == "json" && len(serviceConfig.ResponseFields) > 0 {
			body, _ := io.ReadAll(resp.Body)
			var jsonResp map[string]interface{}
			if err := json.Unmarshal(body, &jsonResp); err == nil {
				if serviceConfig.ErrorField != "" {
					if errMsg, ok := jsonResp[serviceConfig.ErrorField].(string); ok && errMsg != "" {
						result.Valid = false
						result.Message = strings.ToLower(errMsg)
						return result
					}
				}
				
				if serviceConfig.SuccessField != "" {
					if ok, exists := jsonResp[serviceConfig.SuccessField].(bool); exists && ok {
						result.Valid = true
						result.Message = "valid"
						if serviceConfig.DetailsFormat != "" {
							result.Details = renderTemplate(serviceConfig.DetailsFormat, flattenJSON(jsonResp))
						}
					} else {
						result.Valid = false
						result.Message = "invalid key"
					}
				} else {
					flattened := flattenJSON(jsonResp)
					hasData := false
					for _, field := range serviceConfig.ResponseFields {
						if _, exists := flattened[field]; exists {
							hasData = true
							break
						}
					}
					
					if hasData {
						result.Valid = true
						result.Message = "valid"
						if serviceConfig.DetailsFormat != "" {
							result.Details = renderTemplate(serviceConfig.DetailsFormat, flattened)
						}
					} else {
						result.Valid = false
						result.Message = "invalid key"
					}
				}
			} else {
				result.Valid = false
				result.Message = "invalid response format"
			}
		} else {
			result.Valid = true
			result.Message = "valid"
		}
	}

	if resp.StatusCode != serviceConfig.SuccessStatus {
		result.Valid = false
		result.Message = fmt.Sprintf("invalid (http %d)", resp.StatusCode)
	}

	return result
}

func renderTemplate(tmpl string, data map[string]string) string {
	t, err := template.New("tmpl").Parse(tmpl)
	if err != nil {
		return tmpl
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return tmpl
	}
	return buf.String()
}

func flattenJSON(data map[string]interface{}) map[string]string {
	result := make(map[string]string)
	for key, value := range data {
		switch v := value.(type) {
		case string:
			result[key] = v
		case map[string]interface{}:
			for subKey, subValue := range v {
				if str, ok := subValue.(string); ok {
					result[key+"."+subKey] = str
				}
			}
		default:
			result[key] = fmt.Sprintf("%v", v)
		}
	}
	return result
}

func verifyAWS(accessKey, secretKey string, result VerificationResult) VerificationResult {
	if secretKey == "" {
		if strings.HasPrefix(accessKey, "AKIA") && len(accessKey) == 20 {
			result.Valid = false
			result.Message = "key format valid but secret key required"
			result.Details = "use: roq -s aws -k AKIA... -secret YOUR_SECRET_KEY"
		} else {
			result.Valid = false
			result.Message = "invalid aws access key format"
		}
		return result
	}

	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		config.WithRegion("us-east-1"),
	)
	if err != nil {
		result.Valid = false
		result.Message = "failed to create aws config: " + err.Error()
		return result
	}

	resp, err := sts.NewFromConfig(cfg).GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		result.Valid = false
		if strings.Contains(err.Error(), "InvalidClientTokenId") {
			result.Message = "invalid credentials (access key not found)"
		} else if strings.Contains(err.Error(), "SignatureDoesNotMatch") {
			result.Message = "invalid credentials (incorrect secret key)"
		} else {
			result.Message = "verification failed: " + err.Error()
		}
		return result
	}

	result.Valid = true
	result.Message = "valid"
	if resp.Account != nil && resp.Arn != nil {
		result.Details = fmt.Sprintf("account: %s, arn: %s", *resp.Account, *resp.Arn)
	}
	return result
}

func maskKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + strings.Repeat("*", len(key)-8) + key[len(key)-4:]
}
