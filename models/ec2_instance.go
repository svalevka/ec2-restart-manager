// models/ec2_instance.go
package models

type EC2Instance struct {
	AWSAccountName   string `csv:"AWS Account Name"`
	AWSAccountNumber string `csv:"AWS Account Number"`
	State            string `csv:"State"`
	EC2Name          string `csv:"EC2 Name"`
	Service          string `csv:"Service"`
	Owner            string `csv:"Owner"`
	ID               string `csv:"ID"`
	Region           string `csv:"Region"`
	// Add other fields as needed
}

type TemplateData struct {
	Title				  string
	Instances              []EC2Instance
	UniqueOwners           []string
	SelectedOwner          string
	UniqueServices         []string
	SelectedService        string
	UniqueAWSAccountNames  []string
	SelectedAWSAccountName string
	UniqueRegions          []string
	SelectedRegion         string
	IsLoggedIn			   bool
	UserName 			   string	
	AzureAuthenticated	   bool
}

// Example of TemplateData struct:

// data := models.TemplateData{
//     Instances: instances,
//     UniqueOwners: []string{
//         "Alice",
//         "Bob",
//         "Charlie",
//     },
//     SelectedOwner: "", // No owner selected
//     UniqueServices: []string{
//         "Caching Service",
//         "Database Service",
//         "Web Service",
//     },
//     SelectedService: "", // No service selected
//     UniqueAWSAccountNames: []string{
//         "DevAccount",
//         "ProdAccount",
//         "TestAccount",
//     },
//     SelectedAWSAccountName: "", // No AWS account name selected
//     UniqueRegions: []string{
//         "eu-central-1",
//         "us-east-1",
//         "us-west-2",
//     },
//     SelectedRegion: "", // No region selected
// }
