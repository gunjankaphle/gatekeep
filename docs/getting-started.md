# Getting Started with GateKeep

**A practical guide to adopting GateKeep for your company's Snowflake permissions management.**

---

## What is GateKeep?

GateKeep is a GitOps-based tool that manages Snowflake permissions through YAML configuration files. Instead of running manual SQL commands or scripts, you:

1. Define permissions in a YAML file
2. Create a pull request with changes
3. Review the generated SQL automatically
4. Merge the PR → Changes apply to Snowflake automatically

**Key Benefits:**
- ✅ **5-10x faster** than traditional tools (parallel execution)
- ✅ **Preview SQL changes** before applying (via PR comments)
- ✅ **Full audit trail** (Git history + PostgreSQL logs)
- ✅ **Rollback support** (just `git revert`)
- ✅ **No manual commands** needed for daily work

---

## Prerequisites

Before you start, make sure you have:

- [ ] **GitHub repository** for your infrastructure code
- [ ] **Snowflake account** with ACCOUNTADMIN access
- [ ] **PostgreSQL database** for audit logs (AWS RDS, Cloud SQL, or any PostgreSQL instance)
- [ ] **Service account** in Snowflake for GateKeep to use

### Create Snowflake Service Account

```sql
-- In Snowflake, as ACCOUNTADMIN
CREATE USER gatekeep_service_account
  PASSWORD = 'strong-password-here'
  DEFAULT_ROLE = ACCOUNTADMIN
  MUST_CHANGE_PASSWORD = FALSE;

GRANT ROLE ACCOUNTADMIN TO USER gatekeep_service_account;
```

---

## Step 1: Initial Setup (20 minutes)

### 1.1 Create Snowflake Configuration Directory

In your infrastructure repository:

```bash
cd ~/your-company-infrastructure

# Create directory for Snowflake configs
mkdir -p snowflake

# Create your first configuration file
cat > snowflake/prod.yaml <<EOF
version: 1.0

# Define roles
roles:
  - name: DATA_ANALYST
    comment: "Analysts team - read-only access"

  - name: DATA_ENGINEER
    parent_roles: [DATA_ANALYST]
    comment: "Engineers team - read and write access"

# Assign roles to users
users:
  - name: alice@company.com
    roles: [DATA_ANALYST]

  - name: bob@company.com
    roles: [DATA_ENGINEER]

# Grant database permissions
databases:
  - name: ANALYTICS_DB
    schemas:
      - name: PUBLIC
        tables:
          - name: CUSTOMERS
            grants:
              - to_role: DATA_ANALYST
                privileges: [SELECT]
              - to_role: DATA_ENGINEER
                privileges: [SELECT, INSERT, UPDATE, DELETE]

          - name: ORDERS
            grants:
              - to_role: DATA_ANALYST
                privileges: [SELECT]
              - to_role: DATA_ENGINEER
                privileges: [SELECT, INSERT, UPDATE, DELETE]

# Grant warehouse permissions
warehouses:
  - name: ANALYTICS_WH
    grants:
      - to_role: DATA_ANALYST
        privileges: [USAGE]
      - to_role: DATA_ENGINEER
        privileges: [USAGE, OPERATE]
EOF
```

### 1.2 Add GitHub Actions Workflows

```bash
# Create workflows directory
mkdir -p .github/workflows

# Download preview workflow (dry-run on PRs)
curl -o .github/workflows/gatekeep-preview.yml \
  https://raw.githubusercontent.com/gunjankaphle/gatekeep/main/.github/workflows/gatekeep-preview.yml

# Download sync workflow (apply on merge)
curl -o .github/workflows/gatekeep-sync.yml \
  https://raw.githubusercontent.com/gunjankaphle/gatekeep/main/.github/workflows/gatekeep-sync.yml
```

### 1.3 Commit and Push

```bash
git add snowflake/ .github/workflows/
git commit -m "Add GateKeep for Snowflake permissions management"
git push origin main
```

---

## Step 2: Configure GitHub Secrets (5 minutes)

### 2.1 Navigate to Repository Settings

```
Your GitHub Repo → Settings → Secrets and variables → Actions → New repository secret
```

### 2.2 Add Required Secrets

Create the following secrets:

| Secret Name | Example Value | Description |
|-------------|---------------|-------------|
| `SNOWFLAKE_ACCOUNT` | `acme-prod` | Your Snowflake account identifier |
| `SNOWFLAKE_USER` | `gatekeep_service_account` | Service account username |
| `SNOWFLAKE_PASSWORD` | `your-password` | Service account password |
| `SNOWFLAKE_DATABASE` | `ANALYTICS_DB` | Default database |
| `SNOWFLAKE_WAREHOUSE` | `ANALYTICS_WH` | Default warehouse |
| `SNOWFLAKE_ROLE` | `ACCOUNTADMIN` | Role to use (needs admin privileges) |
| `POSTGRES_DSN` | `postgres://user:pass@host:5432/gatekeep` | PostgreSQL connection string for audit logs |

**🎉 Setup Complete! GateKeep is now ready to use.**

---

## Step 3: Daily Usage

### Example 1: Add a New User

**Scenario:** New analyst Sarah joins your team and needs access.

```bash
# 1. Create a feature branch
git checkout -b add-sarah-analyst

# 2. Edit the configuration
vim snowflake/prod.yaml

# Add these lines under the 'users:' section:
  - name: sarah@company.com
    roles: [DATA_ANALYST]

# 3. Commit your changes
git add snowflake/prod.yaml
git commit -m "Add Sarah to DATA_ANALYST role"

# 4. Push and create PR
git push origin add-sarah-analyst
gh pr create --title "Add Sarah to analyst role" --body "Granting analyst access to new team member"
```

**What happens automatically:**

1. ✅ GitHub Actions runs a **dry-run**
2. ✅ Bot **comments on your PR** with the SQL that will be executed:
   ```sql
   CREATE USER "sarah@company.com";
   GRANT ROLE DATA_ANALYST TO USER "sarah@company.com";
   ```
3. ✅ Your team **reviews the SQL**
4. ✅ You **merge the PR**
5. ✅ GitHub Actions **automatically applies** changes to Snowflake
6. ✅ Sarah now has analyst access! ✨

**No manual SQL needed. No CLI commands. Just edit YAML and merge.**

---

### Example 2: Grant Access to a New Table

**Scenario:** You created a new `REVENUE` table and analysts need SELECT access.

```bash
# 1. Create branch
git checkout -b grant-revenue-access

# 2. Edit snowflake/prod.yaml
vim snowflake/prod.yaml

# Add under databases → ANALYTICS_DB → PUBLIC → tables:
          - name: REVENUE
            grants:
              - to_role: DATA_ANALYST
                privileges: [SELECT]
              - to_role: DATA_ENGINEER
                privileges: [SELECT, INSERT, UPDATE, DELETE]

# 3. Commit, push, PR
git add snowflake/prod.yaml
git commit -m "Grant access to REVENUE table"
git push origin grant-revenue-access
gh pr create --title "Grant access to REVENUE table"

# 4. Review SQL preview in PR comments
# 5. Merge → Done!
```

The bot will comment:
```sql
GRANT SELECT ON TABLE ANALYTICS_DB.PUBLIC.REVENUE TO ROLE DATA_ANALYST;
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE ANALYTICS_DB.PUBLIC.REVENUE TO ROLE DATA_ENGINEER;
```

---

### Example 3: Remove a User

**Scenario:** Bob leaves the company. Remove his access.

```bash
# 1. Create branch
git checkout -b remove-bob

# 2. Edit snowflake/prod.yaml - delete these lines:
  - name: bob@company.com
    roles: [DATA_ENGINEER]

# 3. Commit, push, PR
git add snowflake/prod.yaml
git commit -m "Remove Bob's access (offboarded)"
git push origin remove-bob
gh pr create --title "Offboard Bob"

# 4. Review and merge
```

GateKeep will automatically revoke Bob's grants.

---

### Example 4: Create a New Role with Hierarchy

**Scenario:** Create a SENIOR_ANALYST role that inherits DATA_ANALYST permissions.

```yaml
# In snowflake/prod.yaml
roles:
  - name: DATA_ANALYST
    comment: "Base analyst role"

  - name: SENIOR_ANALYST
    parent_roles: [DATA_ANALYST]  # ← Inherits all DATA_ANALYST permissions
    comment: "Senior analysts with additional access"

  - name: DATA_ENGINEER
    parent_roles: [SENIOR_ANALYST]  # ← Can chain roles
    comment: "Engineers inherit senior analyst permissions"

# Then grant additional permissions to SENIOR_ANALYST
databases:
  - name: ANALYTICS_DB
    schemas:
      - name: FINANCE_SCHEMA
        tables:
          - name: SENSITIVE_DATA
            grants:
              - to_role: SENIOR_ANALYST
                privileges: [SELECT]
```

---

## Step 4: Testing Locally (Optional)

Want to test changes before creating a PR? Install the CLI:

### Install GateKeep CLI

**macOS:**
```bash
brew install gatekeep
```

**Linux:**
```bash
curl -L https://github.com/gunjankaphle/gatekeep/releases/latest/download/gatekeep-linux-amd64 -o gatekeep
chmod +x gatekeep
sudo mv gatekeep /usr/local/bin/
```

**Windows:**
```powershell
Invoke-WebRequest -Uri https://github.com/gunjankaphle/gatekeep/releases/latest/download/gatekeep-windows-amd64.exe -OutFile gatekeep.exe
```

### Test Your Config

```bash
# Set environment variables
export SNOWFLAKE_ACCOUNT=your-account
export SNOWFLAKE_USER=your-user
export SNOWFLAKE_PASSWORD=your-password

# Validate configuration
gatekeep validate snowflake/prod.yaml

# Dry-run (preview changes)
gatekeep sync --config snowflake/prod.yaml --dry-run

# If everything looks good, create your PR!
```

---

## Infrastructure Requirements

### What You Need to Run

| Component | Required? | Purpose | Cost |
|-----------|-----------|---------|------|
| **GitHub Repository** | ✅ Yes | Store configs, run workflows | Free |
| **Snowflake Account** | ✅ Yes | Target system | Existing |
| **PostgreSQL Database** | ✅ Yes | Audit logs | ~$20/month |
| **GateKeep API Server** | ❌ Optional | REST API for integrations | $0 (only if needed) |

### PostgreSQL Setup

You can use any PostgreSQL provider:

**AWS RDS:**
```bash
# Create db.t3.micro instance (~$15/month)
# Set POSTGRES_DSN to connection string
```

**Google Cloud SQL:**
```bash
# Create db-f1-micro instance (~$10/month)
```

**Supabase (easiest):**
```bash
# Free tier available!
# 1. Create project at supabase.com
# 2. Copy connection string
# 3. Add to GitHub secrets
```

**Docker (for testing):**
```bash
docker run -d \
  -e POSTGRES_PASSWORD=gatekeep \
  -e POSTGRES_DB=gatekeep \
  -p 5432:5432 \
  postgres:16-alpine
```

---

## How It Works: The GitOps Flow

```
┌─────────────────────────────────────────────────────────┐
│  Developer edits snowflake/prod.yaml                    │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────┐
│  Create Pull Request                                    │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────┐
│  GitHub Actions: Run Dry-Run                            │
│  • Reads current Snowflake state                        │
│  • Compares with YAML config                            │
│  • Generates SQL statements                             │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────┐
│  Bot Comments on PR with SQL Preview                    │
│  "Will execute:                                         │
│   CREATE ROLE ...;                                      │
│   GRANT SELECT ON ...;"                                 │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────┐
│  Team Reviews SQL Changes                               │
│  • Looks correct? Approve PR                            │
│  • Issues found? Request changes                        │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────┐
│  PR Merged to Main                                      │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────┐
│  GitHub Actions: Execute Sync                           │
│  • Applies SQL to Snowflake (parallel execution)        │
│  • Records audit log to PostgreSQL                      │
│  • Posts status to commit                               │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────┐
│  ✅ Changes Applied to Snowflake                        │
│  All operations logged and auditable                    │
└─────────────────────────────────────────────────────────┘
```

---

## Troubleshooting

### PR Comments Not Appearing

**Issue:** GitHub Actions runs but no SQL preview comment appears.

**Solution:**
```bash
# Check workflow logs in GitHub Actions tab
# Ensure SNOWFLAKE_* secrets are set correctly
# Verify service account has ACCOUNTADMIN privileges
```

### Sync Fails After Merge

**Issue:** GitHub Actions fails when trying to apply changes.

**Common causes:**
1. **Invalid credentials** - Check GitHub secrets
2. **Insufficient permissions** - Service account needs ACCOUNTADMIN
3. **Syntax error in YAML** - Validate locally first with `gatekeep validate`
4. **PostgreSQL connection failed** - Check POSTGRES_DSN secret

### Configuration Drift Detected

**Issue:** GateKeep wants to revoke grants you didn't intend to remove.

**Explanation:** GateKeep runs in **strict mode** by default - it enforces that Snowflake exactly matches your YAML.

**Solutions:**
1. **Add missing grants to YAML** if they should stay
2. **Let GateKeep revoke them** if they're unwanted
3. **Switch to additive mode** (only adds, never revokes) by setting `SYNC_MODE=additive` in workflows

---

## Best Practices

### 1. Start Small
```yaml
# Begin with a single role and a few users
roles:
  - name: TEST_ROLE
users:
  - name: yourself@company.com
    roles: [TEST_ROLE]
```

### 2. Use Clear Role Names
```yaml
# Good
- name: DATA_ANALYST_READONLY
- name: DATA_ENGINEER_FULL_ACCESS

# Avoid
- name: ROLE1
- name: TEMP_ACCESS
```

### 3. Document Changes in Commit Messages
```bash
# Good
git commit -m "Grant analysts access to customer_360 table for Q1 reporting"

# Avoid
git commit -m "update config"
```

### 4. Review SQL Carefully
Always check the SQL preview in PR comments before merging. GateKeep is powerful - it will execute exactly what you specify!

### 5. Use Role Hierarchies
```yaml
# Leverage parent_roles to avoid duplication
roles:
  - name: BASE_ROLE
  - name: ADVANCED_ROLE
    parent_roles: [BASE_ROLE]  # Inherits all BASE_ROLE permissions
```

### 6. Keep Configs Organized
```bash
# For multiple environments
snowflake/
├── dev.yaml
├── staging.yaml
└── prod.yaml
```

---

## Migration from Permifrost

Switching from Permifrost? Here's how:

### 1. Export Current State
```bash
# Run Permifrost one last time to ensure Snowflake is in sync
permifrost run --config permifrost.yaml
```

### 2. Convert Config Format
```yaml
# Permifrost format
roles:
  - test_role:
      warehouses:
        - test_warehouse

# GateKeep format (similar but cleaner)
roles:
  - name: TEST_ROLE

warehouses:
  - name: TEST_WAREHOUSE
    grants:
      - to_role: TEST_ROLE
        privileges: [USAGE]
```

### 3. Run GateKeep Dry-Run
```bash
# Should show minimal changes if configs are equivalent
gatekeep sync --config gatekeep.yaml --dry-run
```

### 4. Gradually Switch Over
Keep both tools running in parallel for a week, then fully switch to GateKeep.

---

## FAQ

**Q: Do I need to run any commands daily?**
A: No! Just edit YAML files and create PRs. GateKeep handles everything automatically.

**Q: What if I make a mistake?**
A: Git revert the commit and merge. GateKeep will undo the changes in Snowflake.

**Q: Can I use GateKeep without GitOps?**
A: Yes! Install the CLI and run `gatekeep sync --config prod.yaml` manually. But GitOps is recommended for teams.

**Q: How much does it cost?**
A: GateKeep is free and open-source. You only pay for PostgreSQL (~$20/month) for audit logs.

**Q: Does it work with multiple Snowflake accounts?**
A: Yes! Create separate config files and workflow secrets for each environment (dev, staging, prod).

**Q: Is it secure?**
A: Yes! Credentials are stored as GitHub secrets (encrypted). All operations are logged for audit.

**Q: How fast is it?**
A: 5-10x faster than Permifrost. 1000 operations complete in ~25 seconds vs ~3 minutes.

---

## Next Steps

1. ✅ Complete the initial setup above
2. ✅ Create your first PR to test the workflow
3. ✅ Review the SQL preview
4. ✅ Merge and watch it sync automatically!
5. 📖 Read [YAML Schema Documentation](yaml-schema.md) for advanced configuration
6. 📖 Check [API Documentation](api.md) if you need programmatic access

---

## Support

- 📖 [Full Documentation](../README.md)
- 🐛 [Report Issues](https://github.com/gunjankaphle/gatekeep/issues)
- 💬 [Discussions](https://github.com/gunjankaphle/gatekeep/discussions)

---

**Welcome to GateKeep!** 🎉 Your Snowflake permissions are now code, reviewable, and automated.
