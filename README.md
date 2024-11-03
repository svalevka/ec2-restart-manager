## Description

Web application for restarting EC2 instances from all accounts and regions.
This application does following:
* Authenticate user using Entra and authorize only users from specific User Group
* Connect to S3 bucket in  `shared-prod` account and retrieve inventory EC2 data from CSV file stored in that bucket. 
* Present Webpage with filtering options to select subset of EC2 instances for specific region, account and Owner.
* Restart selected EC2 instances.

## Docker

Build image

```
export TAG="1.0"
export IMAGE="ec2-restart-manager"
docker build -t ec2-restart-manager:${TAG} .
```

## Authentication

### Create Azure app

1. Sign in to the Azure Portal
Go to Azure Portal and sign in with your Azure account.

2. Register a New Application
    Navigate to Azure Active Directory > App registrations.

    Click New registration.

    Name: Enter a name for your application (e.g., EC2 Restart Manager).

    Supported account types: Choose Accounts in this organizational directory only (Single tenant).

    Redirect URI: Set the redirect URI to your application's callback endpoint (e.g., http://ec2-restart-manager.prod.ld.internal/auth/callback).

    Click Register.

3. Configure Authentication Settings
    In your app's registration page, select Authentication.

    Under Platform configurations, ensure Web is added with your redirect URI.

4. Add Client Secret
    Go to Certificates & secrets.

    Click New client secret.

    Description: Provide a description (e.g., AppSecret).

    Expires: Choose an appropriate expiration period.

    Click Add.

    Copy the client secret value now; you won't be able to retrieve it later.

5. Note Down Important IDs
    Application (client) ID: Found on the app's Overview page.

    Directory (tenant) ID: Also on the Overview page.

    You'll need these IDs for configuring your Go application.

6. Configure API Permissions
    Navigate to API permissions.

    Click Add a permission.

    Select Microsoft Graph > Application permissions.

    Search for and select GroupMember.Read.All.

    Click Add permissions.

    Click Grant admin consent for [Your Tenant] and confirm.

7. Expose Group Claims
    Go to Token configuration.

    Click Add groups claim.

    Select Security groups.

    Save your changes.

### Create Azure Group

Create azure group i.e. EC2-Restart-Manager and take note of its id

### Save secret as env var

Provide Azure client secret as env var
```
export AZURE_AD_CLIENT_SECRET="XXXXXXXXXXXXXXXXXXXXXXXXXXXX"
```

## Start app
```
go run main.go
```

## GUI

![Image](./images/instance_manager.png)

