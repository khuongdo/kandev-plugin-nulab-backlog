# Feasibility & Constraints — Clarifying Questions

## Sources

- [desc] Initial description: "feature dựa trên plugin này https://github.com/kdlbs/kandev-plugin-bitbucket hãy tạo 1 plugin tương tự cho kandev (https://github.com/kdlbs/kandev) nhưng dùng cho nulab backlog dựa trên public api ở đây https://developer.nulab.com/docs/backlog/"
- [scope] Workflow-selected scope: `feature`.

Context known from earlier steps: the plugin installs into Kandev like the Bitbucket plugin (Kandev runs the plugin as a child process; no separate cloud infrastructure needed); the Backlog API authenticates with an API key or OAuth 2.0, limits calls per user, and a Backlog space can be on several domains. The questions below establish what can be done with the resources at hand.

---

## Q1. Backlog environment for real testing

The success criteria require running the main flow with a real Backlog space. Which space do you have available for testing?

- A. A paid space with Git and pull requests enabled, usable for testing
- B. A space, but not sure Git is enabled or that I can create repositories/pull requests
- C. None yet, one must be created (for example with a trial)
- D. Not yet defined
- X. Other (please specify)

[Answer]: C

## Q2. Kandev environment for testing

The plugin needs a Kandev server to install and test on. What do you have?

- A. Running Kandev (self-hosted) with admin rights to install plugins
- B. Can set up Kandev locally when needed
- C. None yet
- D. Not yet defined
- X. Other (please specify)

[Answer]: A

## Q3. Backlog domains to support

A Backlog space can be on `backlog.com`, `backlog.jp`, or `backlogtool.com`. Which domains should the plugin support?

- A. All of the domains above (users enter their space's domain)
- B. Only the domain your team uses (state it under Other)
- C. Not yet defined
- X. Other (please specify)

[Answer]: A

## Q4. Backlog sign-in method

The Bitbucket plugin supports both tokens and OAuth. For Backlog, OAuth needs an application registered on Nulab's developer page; an API key is created by the user in Backlog. Which method should the Backlog plugin support?

- A. Only API key in the first release
- B. Both API key and OAuth 2.0 from the first release (like the Bitbucket plugin)
- C. API key first, OAuth added later
- D. Not yet defined
- X. Other (please specify)

[Answer]: B

## Q5. Builder's skills

The Bitbucket plugin uses Go for the backend and TypeScript for the UI. How familiar is the builder of this plugin with those languages?

- A. Familiar with both Go and TypeScript
- B. Familiar with one of them (state it under Other)
- C. Not familiar; will rely heavily on AI to write code
- D. Not yet defined
- X. Other (please specify)

[Answer]: X. typescript (chỉ quen TypeScript)

## Q6. Time and budget

Is there a time or cost limit for the first release?

- A. No hard limit; done when quality is reached
- B. There is a deadline (state the date under Other)
- C. There is a cost limit (for example, no paying for any extra service)
- D. Not yet defined
- X. Other (please specify)

[Answer]: A

## Q7. Compliance and data requirements

The plugin will read issue and pull request content and store Backlog credentials in Kandev. Do any compliance requirements apply?

- A. No specific requirements; only store credentials safely like the Bitbucket plugin
- B. There are internal or legal data rules (state them under Other, for example Japan's personal information protection law, GDPR)
- C. Not yet defined
- X. Other (please specify)

[Answer]: A

## Q8. Release channel

A usable release is a success criterion. Which channel do you want to release through?

- A. This repo's GitHub Release, and also submit a proposal to get into the official Kandev marketplace
- B. Only this repo's GitHub Release (installed by URL or package file)
- C. Not yet defined
- X. Other (please specify)

[Answer]: A

## Consolidated Summary Confirmation

Summary of the answers:

- Q1: C. No Backlog space for testing yet; one must be created
- Q2: A. Running self-hosted Kandev with admin rights to install plugins
- Q3: A. Support all Backlog domains (users enter their space's domain)
- Q4: B. Support both API key and OAuth 2.0 from the first release
- Q5: Other — "typescript": the builder knows TypeScript, not Go
- Q6: A. No hard deadline or cost limit
- Q7: A. No specific compliance requirements; only store credentials safely like the Bitbucket plugin
- Q8: A. Release through GitHub Release and submit a proposal to the official Kandev marketplace

Does this all look correct before I generate the artifact?

- Looks correct
- Request changes

[Answer]: Looks correct
