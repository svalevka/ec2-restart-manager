package main

import (
    "context"
    "encoding/csv"
    "fmt"
    "html/template"
    "io"
    "log"
    "net/http"
    "sort"
    "strings"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/credentials"
    "github.com/aws/aws-sdk-go-v2/service/ec2"
    "github.com/aws/aws-sdk-go-v2/service/s3"
    "github.com/aws/aws-sdk-go-v2/service/sts"
    "github.com/jszwec/csvutil"
)

// EC2Instance represents an EC2 instance with fields matching CSV headers
type EC2Instance struct {
    AWSAccountName   string `csv:"AWS Account Name"`
    AWSAccountNumber string `csv:"AWS Account Number"`
    State            string `csv:"State"`
    EC2Name          string `csv:"EC2 Name"`
    Service          string `csv:"Service"`
    Owner            string `csv:"Owner"`
    ID               string `csv:"ID"`
    Region           string `csv:"Region"`
    // Include other fields as needed
}

// TemplateData holds data passed to the HTML template
type TemplateData struct {
    Instances              []EC2Instance
    UniqueOwners           []string
    SelectedOwner          string
    UniqueServices         []string
    SelectedService        string
    UniqueAWSAccountNames  []string
    SelectedAWSAccountName string
    UniqueRegions          []string
    SelectedRegion         string
}

var (
    cfg                   aws.Config
    ec2Client             *ec2.Client
    stsClient             *sts.Client
    instances             []EC2Instance
    tmpl                  *template.Template
    uniqueOwners          []string
    uniqueServices        []string
    uniqueAWSAccountNames []string
    uniqueRegions         []string
)

func main() {
    var err error

    // Load AWS configuration
    cfg, err = config.LoadDefaultConfig(context.Background())
    if err != nil {
        log.Fatalf("unable to load SDK config, %v", err)
    }

    // Create S3 and STS clients
    s3Client := s3.NewFromConfig(cfg)
    stsClient = sts.NewFromConfig(cfg)

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

    // Extract unique values
    uniqueOwners = getUniqueOwners(instances)
    uniqueServices = getUniqueServices(instances)
    uniqueAWSAccountNames = getUniqueAWSAccountNames(instances)
    uniqueRegions = getUniqueRegions(instances)

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

// getUniqueOwners extracts a sorted list of unique owners from the instances
func getUniqueOwners(instances []EC2Instance) []string {
    ownerSet := make(map[string]struct{})
    for _, instance := range instances {
        owner := strings.TrimSpace(instance.Owner)
        if owner != "" {
            ownerSet[owner] = struct{}{}
        }
    }

    owners := make([]string, 0, len(ownerSet))
    for owner := range ownerSet {
        owners = append(owners, owner)
    }

    sort.Strings(owners)
    return owners
}

// getUniqueServices extracts a sorted list of unique services from the instances
func getUniqueServices(instances []EC2Instance) []string {
    serviceSet := make(map[string]struct{})
    for _, instance := range instances {
        service := strings.TrimSpace(instance.Service)
        if service != "" {
            serviceSet[service] = struct{}{}
        }
    }

    services := make([]string, 0, len(serviceSet))
    for service := range serviceSet {
        services = append(services, service)
    }

    sort.Strings(services)
    return services
}

// getUniqueAWSAccountNames extracts a sorted list of unique AWS Account Names
func getUniqueAWSAccountNames(instances []EC2Instance) []string {
    accountSet := make(map[string]struct{})
    for _, instance := range instances {
        account := strings.TrimSpace(instance.AWSAccountName)
        if account != "" {
            accountSet[account] = struct{}{}
        }
    }

    accounts := make([]string, 0, len(accountSet))
    for account := range accountSet {
        accounts = append(accounts, account)
    }

    sort.Strings(accounts)
    return accounts
}

// getUniqueRegions extracts a sorted list of unique regions from the instances
func getUniqueRegions(instances []EC2Instance) []string {
    regionSet := make(map[string]struct{})
    for _, instance := range instances {
        region := strings.TrimSpace(instance.Region)
        if region != "" {
            regionSet[region] = struct{}{}
        }
    }

    regions := make([]string, 0, len(regionSet))
    for region := range regionSet {
        regions = append(regions, region)
    }

    sort.Strings(regions)
    return regions
}

// indexHandler serves the main page and handles filtering
func indexHandler(w http.ResponseWriter, r *http.Request) {
    filteredInstances := instances
    selectedOwner := ""
    selectedService := ""
    selectedAWSAccountName := ""
    selectedRegion := ""

    // Handle filtering
    if r.Method == http.MethodPost {
        if err := r.ParseForm(); err != nil {
            http.Error(w, "Failed to parse form data", http.StatusBadRequest)
            return
        }
        selectedOwner = r.FormValue("owner")
        selectedService = r.FormValue("service")
        selectedAWSAccountName = r.FormValue("awsAccountName")
        selectedRegion = r.FormValue("region")

        // Apply filters
        filteredInstances = filterInstances(instances, selectedOwner, selectedService, selectedAWSAccountName, selectedRegion)
    }

    // Prepare data for the template
    data := TemplateData{
        Instances:              filteredInstances,
        UniqueOwners:           uniqueOwners,
        SelectedOwner:          selectedOwner,
        UniqueServices:         uniqueServices,
        SelectedService:        selectedService,
        UniqueAWSAccountNames:  uniqueAWSAccountNames,
        SelectedAWSAccountName: selectedAWSAccountName,
        UniqueRegions:          uniqueRegions,
        SelectedRegion:         selectedRegion,
    }

    // Render the template with filtered instances and unique values
    if err := tmpl.Execute(w, data); err != nil {
        http.Error(w, "Failed to render template", http.StatusInternalServerError)
    }
}

// filterInstances filters instances based on selected criteria
func filterInstances(instances []EC2Instance, owner, service, awsAccountName, region string) []EC2Instance {
    var filtered []EC2Instance
    for _, instance := range instances {
        if (owner == "" || instance.Owner == owner) &&
            (service == "" || instance.Service == service) &&
            (awsAccountName == "" || instance.AWSAccountName == awsAccountName) &&
            (region == "" || instance.Region == region) {
            filtered = append(filtered, instance)
        }
    }
    return filtered
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

        // Map of "accountNumber|region" to list of instance IDs
        accountInstances := make(map[string][]string)
        for _, id := range selectedIDs {
            // Find the instance in the list
            var accountNumber string
            var region string
            for _, instance := range instances {
                if instance.ID == id {
                    accountNumber = instance.AWSAccountNumber
                    region = instance.Region
                    break
                }
            }
            if accountNumber == "" {
                fmt.Fprintf(w, "Instance ID %s not found in the data", id)
                return
            }
            key := accountNumber + "|" + region
            accountInstances[key] = append(accountInstances[key], id)
        }

        // Restart instances grouped by account and region
        for key, instanceIDs := range accountInstances {
            parts := strings.Split(key, "|")
            accountNumber := parts[0]
            region := parts[1]
            err := restartInstances(accountNumber, region, instanceIDs)
            if err != nil {
                fmt.Fprintf(w, "Failed to restart instances in account %s: %v", accountNumber, err)
                return
            }
        }

        fmt.Fprintf(w, "Successfully restarted instances: %v", selectedIDs)
    } else {
        http.Redirect(w, r, "/", http.StatusSeeOther)
    }
}

// restartInstances restarts the given EC2 instances in the specified account and region
func restartInstances(accountNumber, region string, instanceIDs []string) error {
    // Assume the terraform-admin role in the target account
    roleArn := fmt.Sprintf("arn:aws:iam::%s:role/terraform-admin", accountNumber)
    stsClient := sts.NewFromConfig(cfg)

    assumeRoleOutput, err := stsClient.AssumeRole(context.Background(), &sts.AssumeRoleInput{
        RoleArn:         aws.String(roleArn),
        RoleSessionName: aws.String("InstanceRestartSession"),
    })
    if err != nil {
        return fmt.Errorf("failed to assume role in account %s: %v", accountNumber, err)
    }

    // Create a new EC2 client with the assumed role credentials
    creds := aws.NewCredentialsCache(
        credentials.NewStaticCredentialsProvider(
            *assumeRoleOutput.Credentials.AccessKeyId,
            *assumeRoleOutput.Credentials.SecretAccessKey,
            *assumeRoleOutput.Credentials.SessionToken,
        ),
    )

    ec2Client := ec2.NewFromConfig(aws.Config{
        Region:      region,
        Credentials: creds,
    })

    // Restart the instances
    input := &ec2.RebootInstancesInput{
        InstanceIds: instanceIDs,
    }

    _, err = ec2Client.RebootInstances(context.Background(), input)
    if err != nil {
        return fmt.Errorf("failed to restart instances: %v", err)
    }
    return nil
}

// HTML template for the webpage with the "Instance ID" field added
const htmlTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>EC2 Instance Manager</title>
    <!-- Bootstrap CSS -->
    <link
        rel="stylesheet"
        href="https://stackpath.bootstrapcdn.com/bootstrap/4.5.2/css/bootstrap.min.css"
        integrity="sha384-jcjb8FfFc67PDXgM3PyyFtZ+1ZmeZ0FbE7aYF0oZ5VYcQ6Z/XzplKd57V0eAd1Gz"
        crossorigin="anonymous"
    >
    <!-- Custom CSS -->
    <style>
        body {
            background-color: #f0f8ff; /* AliceBlue */
        }
        .table thead th {
            background-color: #28a745; /* Green */
            color: #fff;
        }
        .table-striped tbody tr:nth-of-type(odd) {
            background-color: #d1ecf1; /* Light Blue */
        }
        .table-striped tbody tr:nth-of-type(even) {
            background-color: #ffffff; /* White */
        }
        .table-hover tbody tr:hover {
            background-color: #ffeeba; /* Light Yellow */
        }
        .btn-danger {
            background-color: #dc3545;
            border-color: #dc3545;
        }
        .btn-danger:hover {
            background-color: #c82333;
            border-color: #bd2130;
        }
        .btn-primary {
            background-color: #007bff;
            border-color: #007bff;
        }
        .btn-primary:hover {
            background-color: #0069d9;
            border-color: #0062cc;
        }
        .form-inline .form-control {
            width: auto;
        }
        h1 {
            color: #007bff;
        }
    </style>
    <script>
        function submitForm() {
            document.getElementById('filterForm').submit();
        }
    </script>
</head>
<body>
    <!-- Content starts here -->
    <div class="container mt-4">
        <form method="POST" action="/" id="filterForm">
            <div class="table-responsive">
                <table class="table table-striped table-hover table-bordered">
                    <thead>
                        <tr>
                            <th scope="col">Select</th>
                            <th scope="col">
                                AWS Account Name
                                <select name="awsAccountName" class="form-control" onchange="submitForm()">
                                    <option value="">All Accounts</option>
                                    {{range .UniqueAWSAccountNames}}
                                    <option value="{{.}}" {{if eq . $.SelectedAWSAccountName}}selected{{end}}>{{.}}</option>
                                    {{end}}
                                </select>
                            </th>
                            <th scope="col">State</th>
                            <th scope="col">Name</th>
                            <th scope="col">Instance ID</th>
                            <th scope="col">
                                Service
                                <select name="service" class="form-control" onchange="submitForm()">
                                    <option value="">All Services</option>
                                    {{range .UniqueServices}}
                                    <option value="{{.}}" {{if eq . $.SelectedService}}selected{{end}}>{{.}}</option>
                                    {{end}}
                                </select>
                            </th>
                            <th scope="col">
                                Owner
                                <select name="owner" class="form-control" onchange="submitForm()">
                                    <option value="">All Owners</option>
                                    {{range .UniqueOwners}}
                                    <option value="{{.}}" {{if eq . $.SelectedOwner}}selected{{end}}>{{.}}</option>
                                    {{end}}
                                </select>
                            </th>
                            <th scope="col">
                                Region
                                <select name="region" class="form-control" onchange="submitForm()">
                                    <option value="">All Regions</option>
                                    {{range .UniqueRegions}}
                                    <option value="{{.}}" {{if eq . $.SelectedRegion}}selected{{end}}>{{.}}</option>
                                    {{end}}
                                </select>
                            </th>
                        </tr>
                    </thead>
                    <tbody>
                        {{if .Instances}}
                        {{range .Instances}}
                        <tr>
                            <td><input type="checkbox" name="instanceIDs" value="{{.ID}}"></td>
                            <td>{{.AWSAccountName}}</td>
                            <td>{{.State}}</td>
                            <td>{{.EC2Name}}</td>
                            <td>{{.ID}}</td>
                            <td>{{.Service}}</td>
                            <td>{{.Owner}}</td>
                            <td>{{.Region}}</td>
                        </tr>
                        {{end}}
                        {{else}}
                        <tr>
                            <td colspan="8" class="text-center">No instances found for the selected criteria.</td>
                        </tr>
                        {{end}}
                    </tbody>
                </table>
            </div>
            <button type="submit" formaction="/restart" class="btn btn-danger">Restart Selected Instances</button>
        </form>
    </div>
    <!-- Optional JavaScript and dependencies -->
    <script src="https://code.jquery.com/jquery-3.5.1.slim.min.js"
        integrity="sha384-DfXdz2htPH0lsSSs5nCTpuj/zy4C+OGpamoFVy38MVBnE+IbbVYUew+OrCXaRkfj"
        crossorigin="anonymous">
    </script>
    <script src="https://cdn.jsdelivr.net/npm/bootstrap@4.5.2/dist/js/bootstrap.bundle.min.js"
        integrity="sha384-LtrjvnR4GearynP+Usg2lNk9g4+T1F0rSuf6KqjFy9bUazhw+pJiq2P8x2c59ZJP"
        crossorigin="anonymous">
    </script>
</body>
</html>
`
