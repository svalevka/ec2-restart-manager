## Description

Web application for restarting EC2 instances from all accounts and regions.
This application does following:
* Authenticate user using Entra and authorize only users from specific User Group
* Connect to S3 bucket in  `shared-prod` account and retrieve inventory EC2 data from CSV file stored in that bucket. 
* Present Webpage with filtering options to select subset of EC2 instances for specific region, account and Owner.
* Restart selected EC2 instances.

URLs:
* Dev:  https://ec2-restart-manager.dev.ld.internal
* Prod: https://ec2-restart-manager.prod.ld.internal
* Local: http://localhost:8080

## Configuration

Configuration file is stored in `config/config.yml`
Example of using configuration in code: 
```
bucket := cfg.S3.Bucket
key := cfg.S3.Key

```

## Docker

Use [deploy.sh](./deploy.sh) script for an example


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


### **Login Flow**

   - **Login Initiation**:
     - When an unauthenticated user tries to access a protected route, they’re redirected to the `/login` endpoint.
     - The `LoginHandler` then redirects them to the Azure AD login page, passing along the necessary OAuth2 parameters, including the client ID, redirect URI, and scopes.

   - **Authentication with Azure AD**:
     - The user authenticates with Azure AD, which issues an authorization code.
     - Azure AD then redirects the user back to the application’s `/callback` endpoint with this code.

   - **Exchange Authorization Code for Access Token**:
     - The `CallbackHandler` receives the authorization code and exchanges it with Azure AD for an access token.
     - This token is used to retrieve the user’s profile information from Microsoft Graph, particularly their display name.

   - **Create Session and Set Cookie**:
     - A unique session ID is generated for the user, and this session ID, along with the user's display name, is stored in the server-side `SessionStore`.
     - A `session_id` cookie is then set in the user’s browser, containing the session ID, and the user is redirected to the application’s root URL.

   - **Middleware Verification**:
     - For each subsequent request, the `AuthMiddleware` checks the `session_id` cookie in the request and verifies that this session ID exists in `SessionStore`.
     - If valid, the request proceeds; otherwise, the user is redirected back to the `/login` endpoint.

### **Logout Flow**

   - **Logout Handler**:
     - When the user clicks the “Logout” button, they are redirected to the `/logout` endpoint, which is handled by `LogoutHandler`.
   
   - **Session Invalidation**:
     - The `LogoutHandler` removes the user’s session from `SessionStore`, effectively invalidating the session on the server side.
     - The `session_id` cookie in the browser is also cleared by setting its `Max-Age` to `-1` and `Expires` to a past date.

   - **Azure AD Logout**:
     - To fully log the user out, including from Azure AD, the `LogoutHandler` redirects the user to Azure AD’s logout URL.
     - The logout URL includes a `post_logout_redirect_uri` parameter that points to the root URL of the application (dynamically generated based on the request’s `Host` header).
     - After Azure AD clears its session, it redirects the user back to the application’s root URL, where they are now logged out.

## Permissions
App needs to be run with AWS permissions to 
* Read S3 inventory bucket in 'config/config.yml'
* Assume IAM role 'ec2-restart-manager-restarter' in each AWS account. This role must already exist

## IAM Permissions
Application runs in shared-${env} account in EKS using appropriate Service Account role defined in eks terragrunt configuration
In each AWS Account, dedicated IAM Role 'ec2-restart-manager-restarter' is created.
This role can be assumed by the app cross account. This role also has permissions to restart EC2 instances.

## Development

The app can be ran using few methods, depending on where in development cycle you are:
* Run go code directly using your development server
  * Application has access to dev accounts only
* Run it as a container in dev EKS cluster
  * Application has access to dev accounts only
* Run it as a container in Prod EKS cluster
  * Application has access to all accounts

Example development cycle:

* Run code in your workstation as below:
    * Ensure your workstaion has access to list S3 bucket defined in config file
    * Ensure your AWS profile is for `shared-dev` account, i.e. `export AWS_DEFAULT_PROFILE=shared-dev.SharedDevAdministrators`
    * Ensure your workstaion has access to AWS Secrets manager value for secret `platform/ec2-restart-manager` in `shared-dev` AWS account
    * Opet terminal and run  `go run main.go`
      * If you need debug output in console, set env var `export DEBUG=true` before strarting the app
    * Open browser page on http://localhost:8080  
* Run code in `shared-dev` account in EKS
    * Build and push Docker image as per `./deploy.sh`
    * Update Docker image tag in corresponding Helm template `ec2-restart-manager` in Platform team EKS namespace in `shared-dev` account
    * Open browser on https://ec2-restart-manager.dev.ld.internal
* Promote code to `shared-prod` account in EKS.
    * Repeat the same as above for `shared-prod` account
    * Open browser on https://ec2-restart-manager.prod.ld.internal

Three distinct Azure AD apps created to ensure callback addresses are used for each scenario above. See `config/config.yml` for details.

