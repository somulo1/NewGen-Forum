# Seeding

There is no demo seeder. A live deployment starts empty and everything (people, messages, tickets, projects) is created by its
users. The only seeding is reference data, with `cmd/bootstrap`:

1. `go run ./cmd/migrate`
2. `go run ./cmd/bootstrap platform` installs the shipped sounds (the API also does this at start-up).
3. `go run ./cmd/bootstrap owner` (name, email, organization name and web name from `-name -email -org -slug` or `HUB_OWNER_NAME`,
   `HUB_OWNER_EMAIL`, `HUB_ORG_NAME`, `HUB_ORG_SLUG`) creates the owner's account and organization, or, if they exist,
   makes sure the account is active and unlocked and that the person is the organization's active owner. A password is generated
   and printed once (or taken from `HUB_OWNER_PASSWORD`); it must be changed at first sign-in. It never resets an existing password
   unless `-reset-password` is given.
4. After an organization exists (sign-up, then Create organization, or `POST /v1/orgs`):
   `go run ./cmd/bootstrap org -slug <slug> -roles starter` makes sure it has its catalog of pages and rights and adds the
   starter roles: Admin, Head of department, IT support, Finance approver, Staff, Guest and the sector, branch and unit
   managers, with Staff as the default role. Use `-dry-run` first. Leave `-roles` out to give it only the catalog and the Owner
   role, and let it build its own roles on the Roles page.

Everything here is safe to run again: it only adds what is missing and never overwrites what an organization changed. Approval
levels are left at none; each organization decides who approves what. The starter roles are in
`internal/access/starter_roles.json`. Tests use their own sample roles in `internal/testkit/test_roles.json`.
