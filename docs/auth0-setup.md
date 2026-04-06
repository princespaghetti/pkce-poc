# Auth0 Setup Guide

This guide walks through the Auth0 configuration needed to run the PKCE POC. The entire setup takes about 5 minutes.

## Create an Auth0 Account

Sign up for a free account at [auth0.com](https://auth0.com). This creates your first tenant automatically. Your tenant domain will look something like `dev-xxxxxxxx.us.auth0.com`.

If you sign up with Google, you'll already have a Google social connection configured — which means you can use your Google identity to test the login flow without creating a separate database user.

## Create the Application

Navigate to **Applications → Applications → Create Application**.

- **Name**: Whatever you'd like (e.g., `PKCE POC`)
- **Type**: Native

Choose Native because this is a public client — the CLI can't securely store a client secret. Native apps in Auth0 have PKCE enabled by default.

After creation, you'll land on the Quickstart tab showing technology options (Android, iOS, Flutter, etc.). None of these apply to a Go CLI. Skip this tab entirely and go to **Settings**.

### Application Settings

Under the Settings tab, configure the following:

**Application Login URI**: Leave blank. This is for scenarios where Auth0 initiates the login, which doesn't apply here since the CLI always starts the flow.

**Allowed Callback URLs**: `http://localhost:8085/callback`

This is where Auth0 redirects after the user authenticates. The CLI spins up a temporary HTTP server on this port to catch the callback.

**Allowed Logout URLs**: `http://localhost:8085`

This is where Auth0 redirects after the logout flow. The CLI hits Auth0's `/v2/logout` endpoint with a `returnTo` parameter, and Auth0 validates it against this list.

Auth0 may show a banner suggesting you turn off "Non-Verifiable Callback URI End-User Confirmation" to skip an extra security prompt during development. This is fine to disable for a localhost POC.

Save your changes. Note your **Client ID** and **tenant domain** — you'll need both for the CLI configuration.

## Create the API

Navigate to **Applications → APIs → Create API**.

> **Important**: Don't confuse this with the APIs tab on your application's page. You need to create the API from the top-level APIs section in the sidebar.

You'll see the **Auth0 Management API** already listed — this is a system API for managing Auth0 itself. You don't want to use this for your project.

Click **Create API** and fill in:

- **Name**: Whatever you'd like (e.g., `PKCE POC API`)
- **Identifier**: `http://localhost:8080/api`
- **Signing Algorithm**: RS256 (default)

The identifier becomes your `audience` value. It's a logical identifier, not a real URL — Auth0 doesn't try to reach it. It's what the CLI requests during authorization and what the API validates in the JWT's `aud` claim.

### Enable Offline Access (for Refresh Tokens)

If you want the CLI to support token refresh (so users don't need to re-login when the access token expires), two settings are required:

1. On your **API's Settings** tab, enable **Allow Offline Access** under Access Settings. This allows the CLI to request the `offline_access` scope and receive a refresh token alongside the access token.

2. On your **Application's Settings** tab, enable **Refresh Token Rotation**. This is required for the refresh flow to work correctly — Auth0 issues a new refresh token with each use and invalidates the previous one.

### Access Policy

The default access policy settings are fine:

- **Within user access**: Allow via client-grant
- **Within client access**: Allow via client-grant

"Within user access" is what matters for the CLI flow — it allows your Native application to request tokens for this API. "Within client access" is for machine-to-machine flows, which this POC doesn't use.

## Authorize the Application

After creating the API, go back to your application: **Applications → Applications → [your app] → APIs tab**.

You should see your custom API listed. Ensure **User Access** is set to **Authorized**. If it isn't, the CLI will authenticate successfully but Auth0 won't include your API's audience in the token, and the API server will reject it.

This was the one adjustment needed during implementation — everything else worked with defaults.

## What You Don't Need to Configure

- **Client secret**: PKCE replaces it. Auth0 handles this automatically for Native apps.
- **Custom scopes/permissions**: Not needed for the MVP. The API just checks that the token is valid and addressed to the right audience.
- **Custom database users**: If you signed up for Auth0 with Google, the Google social connection is already configured and you can authenticate with your Google identity during testing.

## Configuration Values

After completing the setup, you'll need these values for the CLI and API:

| Value | Where to find it | Used by |
|-------|------------------|---------|
| Tenant domain | Settings tab on your application, or the top-right of the Auth0 dashboard | CLI (authorization URL, token exchange) and API (JWT issuer validation) |
| Client ID | Settings tab on your application | CLI (authorization request, token exchange) |
| API audience/identifier | APIs → your API → General Settings | CLI (audience parameter) and API (JWT audience validation) |

## Verifying the Setup

Once the CLI and API are running, a successful flow looks like:

1. Run the `login` command — your browser opens to Auth0's login page
2. Authenticate with your Google account (or whichever connection you have configured)
3. Auth0 redirects to `http://localhost:8085/callback` — the CLI catches this and exchanges the code for tokens
4. Run the `fetch` command — the CLI sends the token to the API, the API validates it against Auth0's JWKS endpoint, and returns data

If something goes wrong, check the [Auth0 logs](https://manage.auth0.com/#/logs) — they show detailed information about failed authorization and token exchange attempts.
