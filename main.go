package main

import (
    "context"
    "encoding/csv"
    "fmt"
    "html/template"
    "io"
    "log"
    "net/http"
    "strings"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/ec2"
    "github.com/aws/aws-sdk-go-v2/service/s3"
    "github.com/jszwec/csvutil"
)

// EC2Instance represents an EC2 instance with fields matching CSV headers
type EC2Instance struct {
    AWSAccountID       string `csv:"AWS Account ID"`
    AWSAccountName     string `csv:"AWS Account Name"`
    Region             string `csv:"Region"`
    ID                 string `csv:"ID"`
    State              string `csv:"State"`
    EC2Name            string `csv:"EC2 Name"`
    InternalIPAddress  string `csv:"Internal IP Address"`
    ExternalIPAddress  string `csv:"External IP Address"`
    PrivateIPDNSName   string `csv:"Private IP DNS Name"`
    OSVersion          string `csv:"OS Version"`
    NetworkPosition    string `csv:"NetworkPosition"`
    TerragruntPath     string `csv:"TerragruntPath"`
    Product            string `csv:"Product"`
    EnvironmentName    string `csv:"EnvironmentName"`
    ManagedBy          string `csv:"ManagedBy"`
    Description        string `csv:"Description"`
    Name               string `csv:"Name"`
    Service            string `csv:"Service"`
    EnvironmentClass   string `csv:"EnvironmentClass"`
    Owner              string `csv:"Owner"`
    // Add other fields as needed
}

var (
    cfg       aws.Config
    ec2Client *ec2.Client
    instances []EC2Instance
    tmpl      *template.Template
)

func main() {
    var err error

    // Load AWS configuration
    cfg, err = config.LoadDefaultConfig(context.Background())
    if err != nil {
        log.Fatalf("unable to load SDK config, %v", err)
    }

    // Create S3 and EC2 clients
    s3Client := s3.NewFromConfig(cfg)
    ec2Client = ec2.NewFromConfig(cfg)

    // Use your specified S3 bucket and CSV file key
    bucket := "ld-shared-prod-eu-west-2-ec2-inventory"
    key := "ec2_inventory-current.csv"

    // Get CSV content from S3
    csvContent, err := getCSVFromS3(s3Client, bucket, key)
    if err != nil {
        log.Fatalf("failed to retrieve CSV from S3: %v", err)
    }

    // Parse the CSV content into a slice of EC2Instance structs
    instances, err = parseCSVToStruct(csvContent)
    if err != nil {
        log.Fatalf("failed to parse CSV data: %v", err)
    }

    // Parse HTML template
    tmpl = template.Must(template.New("index").Parse(htmlTemplate))

    // Start the web server
    http.HandleFunc("/", indexHandler)
    http.HandleFunc("/restart", restartHandler)

    // Listen on all interfaces
    address := "0.0.0.0:8080"
    log.Printf("Server started at http://%s", address)
    log.Fatal(http.ListenAndServe(address, nil))
}

// getCSVFromS3 retrieves the CSV file from S3 and returns its content as a byte slice
func getCSVFromS3(client *s3.Client, bucket, key string) ([]byte, error) {
    // Retrieve the object from S3
    output, err := client.GetObject(context.Background(), &s3.GetObjectInput{
        Bucket: aws.String(bucket),
        Key:    aws.String(key),
    })
    if err != nil {
        return nil, err
    }
    defer output.Body.Close()

    // Read the content of the object
    content, err := io.ReadAll(output.Body)
    if err != nil {
        return nil, err
    }
    return content, nil
}

// parseCSVToStruct parses the CSV content into a slice of EC2Instance structs
func parseCSVToStruct(csvContent []byte) ([]EC2Instance, error) {
    var instances []EC2Instance
    reader := csv.NewReader(strings.NewReader(string(csvContent)))
    decoder, err := csvutil.NewDecoder(reader)
    if err != nil {
        return nil, fmt.Errorf("error creating CSV decoder: %v", err)
    }

    for {
        var instance EC2Instance
        if err := decoder.Decode(&instance); err == io.EOF {
            break
        } else if err != nil {
            return nil, fmt.Errorf("error decoding CSV: %v", err)
        }
        instances = append(instances, instance)
    }
    return instances, nil
}

// indexHandler serves the main page and handles filtering
func indexHandler(w http.ResponseWriter, r *http.Request) {
    filteredInstances := instances

    // Handle filtering by Owner
    if r.Method == http.MethodPost {
        if err := r.ParseForm(); err != nil {
            http.Error(w, "Failed to parse form data", http.StatusBadRequest)
            return
        }
        owner := r.FormValue("owner")
        if owner != "" {
            filteredInstances = filterByOwner(instances, owner)
        }
    }

    // Render the template with filtered instances
    if err := tmpl.Execute(w, filteredInstances); err != nil {
        http.Error(w, "Failed to render template", http.StatusInternalServerError)
    }
}

// restartHandler handles restarting of selected instances
func restartHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodPost {
        if err := r.ParseForm(); err != nil {
            http.Error(w, "Failed to parse form data", http.StatusBadRequest)
            return
        }
        selectedIDs := r.Form["instanceIDs"]
        if len(selectedIDs) == 0 {
            fmt.Fprintf(w, "No instances selected")
            return
        }

        // Call AWS EC2 API to restart instances
        if err := restartInstances(selectedIDs); err != nil {
            fmt.Fprintf(w, "Failed to restart instances: %v", err)
            return
        }
        fmt.Fprintf(w, "Successfully restarted instances: %v", selectedIDs)
    } else {
        http.Redirect(w, r, "/", http.StatusSeeOther)
    }
}

// filterByOwner filters instances by the given owner
func filterByOwner(instances []EC2Instance, owner string) []EC2Instance {
    var filtered []EC2Instance
    for _, instance := range instances {
        if strings.EqualFold(instance.Owner, owner) {
            filtered = append(filtered, instance)
        }
    }
    return filtered
}

// restartInstances restarts the given EC2 instances
func restartInstances(instanceIDs []string) error {
    input := &ec2.RebootInstancesInput{
        InstanceIds: instanceIDs,
    }

    _, err := ec2Client.RebootInstances(context.Background(), input)
    return err
}

// HTML template for the webpage
const htmlTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>EC2 Instance Manager</title>
</head>
<body>
    <h1>EC2 Instances</h1>
    <form method="POST" action="/">
        <label for="owner">Filter by Owner:</label>
        <input type="text" name="owner" id="owner">
        <input type="submit" value="Filter">
    </form>
    <form method="POST" action="/restart">
        <table border="1">
            <tr>
                <th>Select</th>
                <th>Instance ID</th>
                <th>Name</th>
                <th>Owner</th>
            </tr>
            {{range .}}
            <tr>
                <td><input type="checkbox" name="instanceIDs" value="{{.ID}}"></td>
                <td>{{.ID}}</td>
                <td>{{.EC2Name}}</td>
                <td>{{.Owner}}</td>
            </tr>
            {{end}}
        </table>
        <input type="submit" value="Restart Selected Instances">
    </form>
</body>
</html>
`
