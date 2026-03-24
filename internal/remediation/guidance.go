package remediation

import "github.com/cemililik/leakwatch/pkg/finding"

func init() {
	Register("aws-access-key-id", finding.Remediation{
		Title: "Rotate AWS Access Key",
		Steps: []string{
			"Sign in to the AWS IAM console and locate the compromised access key.",
			"Create a new access key for the same IAM user.",
			"Update all services, applications, and CI/CD pipelines that use the old key.",
			"Deactivate the old access key and monitor CloudTrail for any usage.",
			"Delete the old access key after confirming no remaining usage.",
		},
		DocURL:     "https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html#rotating_access_keys_console",
		ConsoleURL: "https://console.aws.amazon.com/iam/home#/security_credentials",
		Urgency:    "immediate",
		Checklist: []string{
			"Review CloudTrail logs for unauthorized usage of the compromised key.",
			"Check for any resources created or modified by the compromised key.",
			"Notify the security team about the exposure.",
			"Scan the codebase for other occurrences of the same key.",
		},
	})

	Register("github-token", finding.Remediation{
		Title: "Revoke GitHub Token",
		Steps: []string{
			"Go to GitHub Settings > Developer settings > Personal access tokens.",
			"Revoke the compromised token immediately.",
			"Create a new token with the minimum required scopes.",
			"Update all integrations and CI/CD pipelines with the new token.",
		},
		DocURL:     "https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens",
		ConsoleURL: "https://github.com/settings/tokens",
		Urgency:    "immediate",
		Checklist: []string{
			"Review the GitHub audit log for unauthorized actions performed with the token.",
			"Check repository and organization settings for unexpected changes.",
			"Notify the security team about the exposure.",
			"Scan for other repositories that may contain the same token.",
		},
	})

	Register("slack-token", finding.Remediation{
		Title: "Revoke Slack Token",
		Steps: []string{
			"Go to Slack App Management at api.slack.com/apps.",
			"Select the affected application and navigate to OAuth & Permissions.",
			"Revoke the compromised bot or user token.",
			"Reinstall the app to generate a new token.",
			"Update all services using the old token.",
		},
		DocURL:     "https://api.slack.com/authentication/rotation",
		ConsoleURL: "https://api.slack.com/apps",
		Urgency:    "high",
		Checklist: []string{
			"Review Slack access logs for unauthorized message reads or posts.",
			"Check for any data exfiltration from private channels.",
			"Notify the workspace administrators about the exposure.",
			"Scan for other locations where the token may be stored.",
		},
	})

	Register("slack-webhook", finding.Remediation{
		Title: "Regenerate Slack Webhook URL",
		Steps: []string{
			"Go to Slack App Management at api.slack.com/apps.",
			"Select the affected application and navigate to Incoming Webhooks.",
			"Remove the compromised webhook URL.",
			"Create a new webhook URL for the target channel.",
			"Update all services that post via the old webhook URL.",
		},
		DocURL:     "https://api.slack.com/messaging/webhooks",
		ConsoleURL: "https://api.slack.com/apps",
		Urgency:    "high",
		Checklist: []string{
			"Review the target channel for unexpected or malicious messages.",
			"Notify the channel members about potential spam from the compromised webhook.",
			"Scan for other locations where the webhook URL may be stored.",
		},
	})

	Register("stripe-api-key-live", finding.Remediation{
		Title: "Roll Stripe Live API Key",
		Steps: []string{
			"Sign in to the Stripe Dashboard and go to Developers > API keys.",
			"Roll the compromised live secret key to generate a new one.",
			"Update all backend services and payment integrations with the new key.",
			"Monitor the Stripe Dashboard for any unauthorized transactions.",
		},
		DocURL:     "https://docs.stripe.com/keys#rolling-keys",
		ConsoleURL: "https://dashboard.stripe.com/apikeys",
		Urgency:    "immediate",
		Checklist: []string{
			"Review recent Stripe events and logs for unauthorized charges or refunds.",
			"Check for any new connected accounts or transfers.",
			"Notify the security and finance teams about the exposure.",
			"Scan the codebase for other occurrences of the same key.",
		},
	})

	Register("stripe-api-key-test", finding.Remediation{
		Title: "Roll Stripe Test API Key",
		Steps: []string{
			"Sign in to the Stripe Dashboard and go to Developers > API keys.",
			"Roll the compromised test secret key to generate a new one.",
			"Update development and staging environments with the new key.",
		},
		DocURL:     "https://docs.stripe.com/keys#rolling-keys",
		ConsoleURL: "https://dashboard.stripe.com/test/apikeys",
		Urgency:    "medium",
		Checklist: []string{
			"Verify no live data was accessible through the test key.",
			"Update CI/CD pipelines that reference the old test key.",
			"Scan for other locations where the test key may be stored.",
		},
	})

	Register("jwt", finding.Remediation{
		Title: "Rotate JWT Signing Key",
		Steps: []string{
			"Generate a new signing key (symmetric secret or asymmetric key pair).",
			"Deploy the new key to all services that issue or validate tokens.",
			"Invalidate all existing tokens by revoking sessions or using a deny list.",
			"Monitor authentication logs for tokens signed with the old key.",
		},
		DocURL:     "https://datatracker.ietf.org/doc/html/rfc7519",
		ConsoleURL: "",
		Urgency:    "high",
		Checklist: []string{
			"Identify all services and clients that consume the JWT.",
			"Check authentication logs for unauthorized access using the leaked token.",
			"Notify the security team about the exposure.",
			"Verify that token expiration and refresh mechanisms are properly configured.",
		},
	})

	Register("database-connection-string", finding.Remediation{
		Title: "Rotate Database Credentials",
		Steps: []string{
			"Change the database user password immediately via your database management tool.",
			"Update all application connection configurations with the new password.",
			"Restart services that use pooled connections to pick up the new credentials.",
			"Review database access logs for unauthorized connections.",
			"Consider restricting the database user's permissions to the minimum required.",
		},
		DocURL:     "https://cheatsheetseries.owasp.org/cheatsheets/Database_Security_Cheat_Sheet.html",
		ConsoleURL: "",
		Urgency:    "immediate",
		Checklist: []string{
			"Check database audit logs for unauthorized queries or data exports.",
			"Verify that network-level access controls (firewall, security groups) are in place.",
			"Notify the security team and database administrators about the exposure.",
			"Scan for other locations where the connection string may be stored.",
		},
	})

	Register("private-key", finding.Remediation{
		Title: "Regenerate Private Key",
		Steps: []string{
			"Generate a new key pair using a strong algorithm (e.g., Ed25519 or RSA 4096).",
			"Revoke any certificates signed by the compromised key.",
			"Remove the compromised public key from all authorized_keys and trust stores.",
			"Deploy the new public key to all systems and services.",
			"Securely delete the compromised private key from all storage locations.",
		},
		DocURL:     "https://cheatsheetseries.owasp.org/cheatsheets/Key_Management_Cheat_Sheet.html",
		ConsoleURL: "",
		Urgency:    "immediate",
		Checklist: []string{
			"Review SSH and TLS access logs for unauthorized connections using the compromised key.",
			"Check certificate transparency logs if the key was used for TLS.",
			"Notify the security team about the exposure.",
			"Audit all systems where the key had access.",
		},
	})

	Register("generic-api-key", finding.Remediation{
		Title: "Rotate API Key",
		Steps: []string{
			"Identify the service provider associated with the API key.",
			"Sign in to the provider's dashboard and locate the API key management section.",
			"Revoke or regenerate the compromised key.",
			"Update all services and integrations with the new key.",
		},
		DocURL:     "",
		ConsoleURL: "",
		Urgency:    "high",
		Checklist: []string{
			"Review the provider's usage and audit logs for unauthorized API calls.",
			"Notify the security team about the exposure.",
			"Scan the codebase for other occurrences of the same key.",
			"Consider using a secrets manager to avoid embedding keys in source code.",
		},
	})

	Register("openai-api-key", finding.Remediation{
		Title: "Rotate OpenAI API Key",
		Steps: []string{
			"Go to platform.openai.com/api-keys.",
			"Create a new API key.",
			"Update all integrations with the new key.",
			"Delete the old API key.",
			"Check usage logs for unauthorized activity.",
		},
		DocURL:     "https://platform.openai.com/docs/guides/safety-best-practices",
		ConsoleURL: "https://platform.openai.com/api-keys",
		Urgency:    "immediate",
		Checklist: []string{
			"Check billing for unauthorized usage.",
			"Review API logs for suspicious activity.",
			"Notify the team about the exposure.",
		},
	})

	Register("anthropic-api-key", finding.Remediation{
		Title: "Rotate Anthropic API Key",
		Steps: []string{
			"Go to console.anthropic.com/settings/keys.",
			"Create a new API key.",
			"Update all integrations with the new key.",
			"Delete the old API key.",
		},
		DocURL:     "https://docs.anthropic.com/en/docs/initial-setup",
		ConsoleURL: "https://console.anthropic.com/settings/keys",
		Urgency:    "immediate",
		Checklist: []string{
			"Check usage logs for unauthorized activity.",
			"Review billing for unexpected charges.",
			"Notify the team about the exposure.",
		},
	})

	Register("gitlab-pat", finding.Remediation{
		Title: "Revoke GitLab Personal Access Token",
		Steps: []string{
			"Go to GitLab Settings > Access Tokens.",
			"Revoke the compromised token.",
			"Create a new token with minimal scopes.",
			"Update CI/CD pipelines with the new token.",
		},
		DocURL:     "https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html",
		ConsoleURL: "https://gitlab.com/-/user_settings/personal_access_tokens",
		Urgency:    "immediate",
		Checklist: []string{
			"Check repository activity for unauthorized changes.",
			"Review CI/CD jobs for suspicious runs.",
			"Audit access logs for unauthorized access.",
		},
	})

	Register("sendgrid-api-key", finding.Remediation{
		Title: "Rotate SendGrid API Key",
		Steps: []string{
			"Go to SendGrid Settings > API Keys.",
			"Create a new key with minimal permissions.",
			"Update email service configuration with the new key.",
			"Delete the old API key.",
		},
		DocURL:     "https://docs.sendgrid.com/ui/account-and-settings/api-keys",
		ConsoleURL: "https://app.sendgrid.com/settings/api_keys",
		Urgency:    "immediate",
		Checklist: []string{
			"Check sent email logs for abuse.",
			"Monitor bounce and spam rates for anomalies.",
			"Notify the team about the exposure.",
		},
	})

	Register("npm-token", finding.Remediation{
		Title: "Revoke NPM Access Token",
		Steps: []string{
			"Run `npm token revoke <token>` or go to npmjs.com > Access Tokens.",
			"Create a new token.",
			"Update CI/CD pipelines with the new token.",
			"Check for unauthorized package publishes.",
		},
		DocURL:     "https://docs.npmjs.com/about-access-tokens",
		ConsoleURL: "https://www.npmjs.com/settings/~/tokens",
		Urgency:    "immediate",
		Checklist: []string{
			"Check for unauthorized package publishes.",
			"Review download stats for anomalies.",
			"Run npm audit on affected packages.",
		},
	})

	Register("datadog-api-key", finding.Remediation{
		Title: "Rotate Datadog API Key",
		Steps: []string{
			"Go to Datadog Organization Settings > API Keys.",
			"Create a new API key.",
			"Update DD_API_KEY in all deployments.",
			"Delete the old API key.",
		},
		DocURL:     "https://docs.datadoghq.com/account_management/api-app-keys/",
		ConsoleURL: "https://app.datadoghq.com/organization-settings/api-keys",
		Urgency:    "immediate",
		Checklist: []string{
			"Check Datadog audit trail.",
			"Review metric submissions.",
			"Notify SRE team.",
		},
	})

	Register("discord-bot-token", finding.Remediation{
		Title: "Reset Discord Bot Token",
		Steps: []string{
			"Go to Discord Developer Portal.",
			"Select your application.",
			"Navigate to Bot settings.",
			"Click Reset Token.",
			"Update all bot configurations.",
		},
		DocURL:     "https://discord.com/developers/docs/reference",
		ConsoleURL: "https://discord.com/developers/applications",
		Urgency:    "immediate",
		Checklist: []string{
			"Check bot activity logs.",
			"Review server permissions.",
			"Notify server admins.",
		},
	})

	Register("redis-connection-string", finding.Remediation{
		Title: "Rotate Redis Credentials",
		Steps: []string{
			"Connect to Redis.",
			"Create new user/password with ACL SETUSER.",
			"Update all application configs.",
			"Remove old credentials with ACL DELUSER.",
		},
		DocURL:     "https://redis.io/docs/latest/operate/oss_and_stack/management/security/acl/",
		ConsoleURL: "",
		Urgency:    "immediate",
		Checklist: []string{
			"Check Redis MONITOR for unauthorized access.",
			"Review connected clients.",
			"Flush suspicious sessions.",
		},
	})

	Register("snowflake-credentials", finding.Remediation{
		Title: "Rotate Snowflake Password",
		Steps: []string{
			"Log in to Snowflake.",
			"ALTER USER to change password.",
			"Update all JDBC connection strings.",
			"Revoke active sessions with ALTER USER ABORT ALL QUERIES.",
		},
		DocURL:     "https://docs.snowflake.com/en/sql-reference/sql/alter-user",
		ConsoleURL: "https://app.snowflake.com",
		Urgency:    "immediate",
		Checklist: []string{
			"Review QUERY_HISTORY for unauthorized access.",
			"Check ACCESS_HISTORY.",
			"Notify data team.",
		},
	})

	Register("telegram-bot-token", finding.Remediation{
		Title: "Revoke Telegram Bot Token",
		Steps: []string{
			"Open Telegram.",
			"Message @BotFather.",
			"Use /revoke command.",
			"Select the bot.",
			"Create new token with /token.",
			"Update integrations.",
		},
		DocURL:     "https://core.telegram.org/bots/api",
		ConsoleURL: "https://t.me/BotFather",
		Urgency:    "immediate",
		Checklist: []string{
			"Check bot message history.",
			"Review webhook configurations.",
			"Notify team.",
		},
	})

	// Sprint 2 detectors

	Register("huggingface-token", finding.Remediation{
		Title: "Revoke Hugging Face Token",
		Steps: []string{
			"Go to huggingface.co/settings/tokens.",
			"Delete the compromised token.",
			"Create a new token.",
		},
		DocURL:     "https://huggingface.co/docs/hub/security-tokens",
		ConsoleURL: "",
		Urgency:    "immediate",
	})

	Register("deepseek-api-key", finding.Remediation{
		Title: "Rotate DeepSeek API Key",
		Steps: []string{
			"Go to platform.deepseek.com, API Keys section.",
			"Delete the compromised key.",
			"Create a new API key.",
		},
		DocURL:     "https://platform.deepseek.com/api-docs",
		ConsoleURL: "",
		Urgency:    "immediate",
	})

	Register("gcp-service-account", finding.Remediation{
		Title: "Rotate GCP Service Account Key",
		Steps: []string{
			"Go to GCP Console > IAM > Service Accounts.",
			"Delete the compromised key.",
			"Create a new key.",
			"Update all deployments with the new key.",
		},
		DocURL:     "https://cloud.google.com/iam/docs/keys-create-delete",
		ConsoleURL: "https://console.cloud.google.com/iam-admin/serviceaccounts",
		Urgency:    "immediate",
		Checklist: []string{
			"Check Cloud Audit Logs.",
			"Review IAM permissions.",
			"Notify security team.",
		},
	})

	Register("azure-storage-key", finding.Remediation{
		Title: "Rotate Azure Storage Access Key",
		Steps: []string{
			"Go to Azure Portal > Storage Account > Access Keys.",
			"Rotate the compromised key.",
			"Update all connection strings.",
		},
		DocURL:     "https://learn.microsoft.com/en-us/azure/storage/common/storage-account-keys-manage",
		ConsoleURL: "",
		Urgency:    "immediate",
		Checklist: []string{
			"Check Storage Analytics logs.",
			"Review SAS tokens derived from key.",
		},
	})

	Register("azure-entra-secret", finding.Remediation{
		Title: "Rotate Azure Entra ID Client Secret",
		Steps: []string{
			"Go to Azure Portal > App Registrations > Certificates & Secrets.",
			"Delete the old secret.",
			"Create a new client secret.",
			"Update all applications with the new secret.",
		},
		DocURL:     "https://learn.microsoft.com/en-us/entra/identity-platform/howto-create-service-principal-portal",
		ConsoleURL: "",
		Urgency:    "immediate",
	})

	Register("okta-api-token", finding.Remediation{
		Title: "Revoke Okta API Token",
		Steps: []string{
			"Go to Okta Admin Console > Security > API > Tokens.",
			"Revoke the compromised token.",
		},
		DocURL:     "https://developer.okta.com/docs/guides/create-an-api-token",
		ConsoleURL: "",
		Urgency:    "immediate",
		Checklist: []string{
			"Review system log for unauthorized API calls.",
		},
	})

	Register("twilio-api-key", finding.Remediation{
		Title: "Rotate Twilio API Key",
		Steps: []string{
			"Go to Twilio Console > API Keys.",
			"Delete the compromised key.",
			"Create a new API key.",
		},
		DocURL:     "https://www.twilio.com/docs/iam/api-keys",
		ConsoleURL: "https://console.twilio.com/us1/account/keys-credentials/api-keys",
		Urgency:    "immediate",
		Checklist: []string{
			"Check call/SMS logs for abuse.",
		},
	})

	Register("mailgun-api-key", finding.Remediation{
		Title: "Rotate Mailgun API Key",
		Steps: []string{
			"Go to Mailgun Dashboard > API Keys.",
			"Create a new key.",
			"Update all integrations.",
			"Delete the old key.",
		},
		DocURL:     "https://documentation.mailgun.com/docs/mailgun/api-reference/authentication/",
		ConsoleURL: "",
		Urgency:    "immediate",
		Checklist: []string{
			"Check sending logs for unauthorized emails.",
		},
	})

	Register("hashicorp-vault-token", finding.Remediation{
		Title: "Revoke Vault Token",
		Steps: []string{
			"Run `vault token revoke <token>` or revoke via API.",
		},
		DocURL:     "https://developer.hashicorp.com/vault/docs/commands/token/revoke",
		ConsoleURL: "",
		Urgency:    "immediate",
		Checklist: []string{
			"Check Vault audit logs.",
			"Review token policies.",
		},
	})

	Register("grafana-api-key", finding.Remediation{
		Title: "Revoke Grafana Service Account Token",
		Steps: []string{
			"Go to Grafana > Administration > Service Accounts.",
			"Delete the compromised token.",
		},
		DocURL:     "https://grafana.com/docs/grafana/latest/administration/service-accounts/",
		ConsoleURL: "",
		Urgency:    "high",
	})

	Register("pagerduty-api-key", finding.Remediation{
		Title: "Rotate PagerDuty API Key",
		Steps: []string{
			"Go to PagerDuty > My Profile > User Settings > API Access.",
			"Create a new key.",
			"Delete the old key.",
		},
		DocURL:     "https://support.pagerduty.com/docs/api-access-keys",
		ConsoleURL: "",
		Urgency:    "high",
	})

	Register("circleci-token", finding.Remediation{
		Title: "Revoke CircleCI Token",
		Steps: []string{
			"Go to CircleCI > User Settings > Personal API Tokens.",
			"Delete the compromised token.",
		},
		DocURL:     "https://circleci.com/docs/managing-api-tokens/",
		ConsoleURL: "",
		Urgency:    "high",
	})

	Register("github-oauth-token", finding.Remediation{
		Title: "Revoke GitHub OAuth Token",
		Steps: []string{
			"Go to GitHub Settings > Developer Settings > OAuth Apps.",
			"Revoke the compromised token.",
		},
		DocURL:     "https://docs.github.com/en/apps/oauth-apps",
		ConsoleURL: "",
		Urgency:    "immediate",
		Checklist: []string{
			"Check GitHub audit log.",
			"Review app permissions.",
		},
	})

	// Sprint 3 detectors

	Register("pypi-api-token", finding.Remediation{
		Title: "Revoke PyPI API Token",
		Steps: []string{
			"Go to pypi.org > Account Settings > API Tokens.",
			"Delete the compromised token.",
			"Create a new scoped token with minimal permissions.",
			"Update CI/CD pipelines with the new token.",
		},
		DocURL:     "https://pypi.org/help/#apitoken",
		ConsoleURL: "https://pypi.org/manage/account/#api-tokens",
		Urgency:    "immediate",
	})

	Register("rubygems-api-key", finding.Remediation{
		Title: "Revoke RubyGems API Key",
		Steps: []string{
			"Go to rubygems.org > Settings > API Keys.",
			"Delete the compromised key.",
			"Create a new API key with minimal scopes.",
			"Update CI/CD pipelines with the new key.",
		},
		DocURL:     "https://guides.rubygems.org/api-key-scopes/",
		ConsoleURL: "https://rubygems.org/settings/edit",
		Urgency:    "immediate",
		Checklist: []string{
			"Check for unauthorized gem publishes.",
		},
	})

	Register("dockerhub-pat", finding.Remediation{
		Title: "Revoke Docker Hub PAT",
		Steps: []string{
			"Go to hub.docker.com > Account Settings > Security.",
			"Delete the compromised personal access token.",
			"Create a new token with appropriate permissions.",
			"Update all Docker CLI and CI/CD configurations.",
		},
		DocURL:     "https://docs.docker.com/security/for-developers/access-tokens/",
		ConsoleURL: "https://hub.docker.com/settings/security",
		Urgency:    "immediate",
		Checklist: []string{
			"Check image push history for unauthorized publishes.",
		},
	})

	Register("digitalocean-token", finding.Remediation{
		Title: "Revoke DigitalOcean Token",
		Steps: []string{
			"Go to cloud.digitalocean.com > API > Tokens.",
			"Delete the compromised token.",
			"Create a new token with required scopes.",
			"Update all integrations with the new token.",
		},
		DocURL:     "https://docs.digitalocean.com/reference/api/",
		ConsoleURL: "https://cloud.digitalocean.com/account/api/tokens",
		Urgency:    "immediate",
		Checklist: []string{
			"Check droplet/resource activity for unauthorized changes.",
		},
	})

	Register("heroku-api-key", finding.Remediation{
		Title: "Regenerate Heroku API Key",
		Steps: []string{
			"Go to dashboard.heroku.com > Account Settings.",
			"Regenerate the API key.",
			"Update all CLI sessions and CI/CD pipelines with the new key.",
		},
		DocURL:     "https://devcenter.heroku.com/articles/authentication",
		ConsoleURL: "https://dashboard.heroku.com/account",
		Urgency:    "immediate",
	})

	Register("vercel-token", finding.Remediation{
		Title: "Revoke Vercel Token",
		Steps: []string{
			"Go to vercel.com > Settings > Tokens.",
			"Delete the compromised token.",
			"Create a new token.",
			"Update all integrations with the new token.",
		},
		DocURL:     "https://vercel.com/docs/rest-api",
		ConsoleURL: "https://vercel.com/account/tokens",
		Urgency:    "high",
	})

	Register("newrelic-api-key", finding.Remediation{
		Title: "Delete New Relic API Key",
		Steps: []string{
			"Go to one.newrelic.com > API Keys.",
			"Delete the compromised key.",
			"Create a new API key.",
			"Update all integrations with the new key.",
		},
		DocURL:     "https://docs.newrelic.com/docs/apis/intro-apis/new-relic-api-keys/",
		ConsoleURL: "https://one.newrelic.com/api-keys",
		Urgency:    "high",
	})

	Register("sentry-token", finding.Remediation{
		Title: "Revoke Sentry Auth Token",
		Steps: []string{
			"Go to sentry.io > Settings > Auth Tokens.",
			"Delete the compromised token.",
			"Create a new auth token.",
			"Update all integrations with the new token.",
		},
		DocURL:     "https://docs.sentry.io/api/auth/",
		ConsoleURL: "https://sentry.io/settings/account/api/auth-tokens/",
		Urgency:    "high",
	})

	Register("shopify-access-token", finding.Remediation{
		Title: "Rotate Shopify Access Token",
		Steps: []string{
			"Go to Shopify Admin > Apps > Manage private apps.",
			"Rotate the compromised access token.",
			"Update all integrations with the new token.",
		},
		DocURL:     "https://shopify.dev/docs/apps/auth",
		ConsoleURL: "",
		Urgency:    "immediate",
		Checklist: []string{
			"Check order/customer data access for unauthorized activity.",
		},
	})

	Register("supabase-service-key", finding.Remediation{
		Title: "Rotate Supabase Service Key",
		Steps: []string{
			"Go to app.supabase.com > Project Settings > API.",
			"Regenerate the service role key.",
			"Update all backend services with the new key.",
		},
		DocURL:     "https://supabase.com/docs/guides/api",
		ConsoleURL: "https://app.supabase.com",
		Urgency:    "immediate",
	})

	Register("cloudflare-api-token", finding.Remediation{
		Title: "Revoke Cloudflare API Token",
		Steps: []string{
			"Go to dash.cloudflare.com > My Profile > API Tokens.",
			"Delete the compromised token.",
			"Create a new token with minimal permissions.",
			"Update all integrations with the new token.",
		},
		DocURL:     "https://developers.cloudflare.com/api/tokens/create/",
		ConsoleURL: "https://dash.cloudflare.com/profile/api-tokens",
		Urgency:    "immediate",
		Checklist: []string{
			"Check DNS changes and firewall rules for unauthorized modifications.",
		},
	})

	Register("notion-token", finding.Remediation{
		Title: "Revoke Notion Integration Token",
		Steps: []string{
			"Go to notion.so > Settings > Connections > Develop or manage integrations.",
			"Revoke the compromised integration token.",
			"Create a new integration secret.",
			"Update all integrations with the new token.",
		},
		DocURL:     "https://developers.notion.com/docs/authorization",
		ConsoleURL: "https://www.notion.so/my-integrations",
		Urgency:    "high",
	})

	Register("linear-api-key", finding.Remediation{
		Title: "Revoke Linear API Key",
		Steps: []string{
			"Go to linear.app > Settings > API.",
			"Delete the compromised key.",
			"Create a new API key.",
			"Update all integrations with the new key.",
		},
		DocURL:     "https://developers.linear.app/docs/graphql/working-with-the-graphql-api",
		ConsoleURL: "https://linear.app/settings/api",
		Urgency:    "high",
	})

	Register("figma-pat", finding.Remediation{
		Title: "Revoke Figma PAT",
		Steps: []string{
			"Go to figma.com > Settings > Personal Access Tokens.",
			"Delete the compromised token.",
			"Create a new personal access token.",
			"Update all integrations with the new token.",
		},
		DocURL:     "https://www.figma.com/developers/api#access-tokens",
		ConsoleURL: "https://www.figma.com/settings",
		Urgency:    "high",
	})

	Register("airtable-pat", finding.Remediation{
		Title: "Revoke Airtable PAT",
		Steps: []string{
			"Go to airtable.com > Account > Developer hub > Personal access tokens.",
			"Delete the compromised token.",
			"Create a new personal access token with minimal scopes.",
			"Update all integrations with the new token.",
		},
		DocURL:     "https://airtable.com/developers/web/guides/personal-access-tokens",
		ConsoleURL: "https://airtable.com/create/tokens",
		Urgency:    "high",
	})
}
