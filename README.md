<div align="center">
  <img src="https://github.com/user-attachments/assets/f30ad220-ea69-45df-ab64-8a12652aa6c6" alt="roq" width="955">
</div>

<br>
<br>
<br>

> [!NOTE] 
> **Test the exposed API keys you got while hunting.**

<br>

- <sub> **supports 76+ services** (AWS, GitHub, Stripe, Slack, OpenAI, and many more) </sub>
- <sub> **validates keys instantly** with proper authentication methods </sub>
- <sub> **extracts user/account details** from valid keys </sub>
- <sub> **rotates random User-Agent** for each request </sub>
- <sub> **clean and pipe friendly output** with JSON support </sub>

<br>
<br>

<h4>Installation</h4>

```bash
go install github.com/1hehaq/roq@latest
```

<br>

<h6>setup autocompletion for service names</h6>

```bash
echo -e "complete -W '\$(roq -list | grep -oP \"(?<=• )[a-z]+\")' roq" >> ~/.bashrc && source ~/.bashrc
```

- <sub>**then try this**</sub>
  ```bash
  roq -s <TAB>
  ```
  ```bash
  roq -s git<TAB>
  ```

<br>
<br>

<h4>Flags</h4>

<pre>
  -s      : service type (required)
  -k      : api key to verify (required)
  -secret : secret key (required for aws, twilio, razorpay, trello)
  -json   : output in json format
  -list   : list all supported services
  -v      : verbose output
  -h      : show help message
</pre>

<br>
<br>

<h4>Example Commands</h4>

```bash
# verify a github token
roq -s github -k ghp_xxxxxxxxxxxx
```

<br>

```bash
# verify aws credentials
roq -s aws -k AKIA... -secret YOUR_SECRET_KEY
```

<br>

```bash
# verify stripe key and get json output
roq -s stripe -k sk_live_xxxxxxxxxxxx -json
```

<br>

```bash
# verify slack token and extract user details
roq -s slack -k xoxb-xxxxxxxxxxxx
```

<br>

```bash
# list all supported services
roq -list
```

<br>

```bash
# pipe multiple keys for batch verification
cat keys.txt | while read key; do roq -s github -k $key -json; done | jq -r 'select(.valid==true)'
```

<br>
<br>

<h4>Adding Custom Services</h4>

<sub>**roq** supports custom service configurations via the `services.yaml` file. You can add your own API services by defining them in the configuration file.</sub>

<br>

**Configuration Location:**
- <sub>Default: `services.yaml` in the current directory</sub>
- <sub>Or specify with environment variable or custom path</sub>

<br>

**Basic Service Structure:**

<img width="2846" height="1580" alt="services YAML reference" src="https://github.com/user-attachments/assets/3d101e34-ba04-4390-8295-3b9881b44e16" />

<details>
<summary><sub><strong>View YAML Code</strong></sub></summary>

<br>

```yaml
services:
  github:
    name: GitHub
    method: GET                         # HTTP method (GET, POST, etc.)
    url: https://api.github.com/user    # API endpoint
    headers:
      Authorization: "token {{.Key}}"   # {{.Key}} is replaced with the API key
      User-Agent: "{{.UserAgent}}"      # user agent string
    success_status: 200                 # expected HTTP status for success
    response_type: json                 # response format (json, xml, etc.)
    response_fields:                    # fields to extract from response
      - login
      - name
    details_format: "user: {{.login}}"  # format for displaying details
    error_field: message                # field containing error message
    requires_secret: false              # whether additional secret is needed
```

</details>

<br>

**Advanced Options:**
- <sub>**Basic Auth**: Use `auth_type: basic`, `auth_user`, and `auth_pass`</sub>
- <sub>**Multiple Secrets**: Set `requires_secret: true` and `secret_name`</sub>
- <sub>**Dynamic URLs**: Use placeholders like `{{.Domain}}` or `{{.Instance}}`</sub>
- <sub>**Custom Success Field**: Define `success_field` for boolean validation</sub>

<br>

<sub>See the [services.yaml](services.yaml) file for more examples of different authentication methods and configurations.</sub>

<br>
<br>

<h4>Supported Services</h4>

<details>
<summary><strong>Click to expand (76 services)</strong></summary>

<br>

**Cloud & Infrastructure**
- AWS, DigitalOcean, GoogleCloud, Heroku, Terraform, Cloudflare, MongoDB, Supabase

**Development & CI/CD**
- GitHub, GitLab, Bitbucket, CircleCI, Buildkite, JFrog, NPM

**Communication**
- Slack, Discord, Telegram, Twilio, Telnyx

**Payment & Commerce**
- Stripe, PayPal, Paystack, Razorpay, Square, Shopify

**Marketing & CRM**
- HubSpot, Mailchimp, SendGrid, Mailgun, MailerLite, SendinBlue, Klaviyo, Omnisend, GetResponse

**Project Management**
- Jira, Trello, Asana, Linear, PagerDuty

**Analytics & Monitoring**
- Datadog, Sentry, PostHog, Grafana, Honeycomb, SonarCloud

**AI & ML**
- OpenAI, HuggingFace, NVIDIA

**Design & Collaboration**
- Figma, Notion, Airtable, Typeform

**Security & DevOps**
- Snyk, OpsGenie, LaunchDarkly, Doppler

**Other Services**
- Algolia, Bitly, Clerk, Eventbrite, Postman, Pulumi, Pushbullet, RabbitMQ, Salesforce, Shodan, Yousign, Zendesk

</details>

<br>
<br>

- **If you see errors or invalid results**
  - <sub> **verify your API key format** </sub>
  - <sub> **check your internet connection** </sub>
  - <sub> **some services require additional parameters (domain, instance, etc.)** </sub>
  - <sub> **use `-v` for verbose output** </sub>
  - <sub> **use `-h` for guidance** </sub>

<br>
<br>

> [!CAUTION] 
> **never use `roq` for any illegal activities. I'm not responsible for your deeds with it. Use responsibly and only on authorized targets.**

<br>
<br>
<br>

<h6 align="center">kindly for hackers</h6>

<div align="center">
  <a href="https://github.com/1hehaq"><img src="https://img.icons8.com/material-outlined/20/808080/github.png" alt="GitHub"></a>
  <a href="https://twitter.com/1hehaq"><img src="https://img.icons8.com/material-outlined/20/808080/twitter.png" alt="X"></a>
</div>
