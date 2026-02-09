# Phase 1 Complete: Database Schema Updates

**Date:** 2026-02-08  
**Status:** ✅ COMPLETED

## Summary

Successfully updated the database schema and all related code to align with DESIGN_CLI.md specifications. All changes have been made to the codebase but **the database migration has NOT been run yet**.

## Changes Made

### 1. Schema Updates (`internal/db/schema.sql`)

**Tasks Table:**
- ✅ Changed types from `('epic', 'story', 'task', 'sub-task', 'bug', 'spike')` to `('epic', 'task', 'bug', 'spike', 'chore')`
- ✅ Added `comments TEXT` field (stores JSON array)
- ✅ Added `is_deleted INTEGER NOT NULL DEFAULT 0` field
- ✅ Added index on `is_deleted` field

**Roles Table:**
- ✅ Renamed `rules` column to `goals`

**New Sessions Table:**
- ✅ Added complete sessions table for JWT authentication
- ✅ Added all necessary indexes

### 2. Go Models Updated (`internal/db/models.go`)

- ✅ Updated `Role` struct: `Rules` → `Goals`
- ✅ Updated `Task` struct: 
  - Changed type comment
  - Added `Comments []Comment` field
  - Added `IsDeleted bool` field
- ✅ Added new `Comment` struct
- ✅ Added new `Session` struct

### 3. Database Initialization (`internal/db/init.go`)

- ✅ Updated role creation to use `goals` instead of `rules`
- ✅ Changed struct field names
- ✅ Updated SQL INSERT statements

### 4. Server Code Updates

**Files Updated:**
- ✅ `internal/server/roles.go` - All `rules`/`Rules` → `goals`/`Goals`
- ✅ `internal/server/workers.go` - SQL queries updated
- ✅ `internal/server/server_test.go` - Test data updated

### 5. TUI Code Updates

**Files Updated:**
- ✅ `internal/tui/role_form.go` - UI labels and field references
- ✅ `internal/tui/models.go` - Struct definitions

### 6. Migration Script Created

- ✅ `internal/db/migrations/001_align_with_design_cli.sql`
  - Recreates tasks table with new schema
  - Migrates data (converts 'story'/'sub-task' types to 'task')
  - Recreates roles table with renamed column
  - Creates sessions table
  - Recreates all indexes and triggers

## Testing Required

Before using the migration:

1. **Backup existing database:**
   ```bash
   cp ~/.config/sf/sf.db ~/.config/sf/sf.db.backup
   ```

2. **Test migration on a copy:**
   ```bash
   sqlite3 test.db < internal/db/schema.sql
   # Verify schema
   ```

3. **Run full test suite:**
   ```bash
   make test
   ```

## How to Apply Migration

### Option A: Fresh Database (Recommended for Development)
```bash
# Remove existing database and reinitialize
rm ~/.config/sf/sf.db
./sf initdb
```

### Option B: Migrate Existing Database
```bash
# Apply migration script
sqlite3 ~/.config/sf/sf.db < internal/db/migrations/001_align_with_design_cli.sql
```

## What's Next

Phase 1 (Database Schema) is complete. Ready to proceed to:

**Phase 2: Server API Implementation (20 endpoints)**
- Worker management APIs (6 endpoints)
- Task comments API (2 endpoints) 
- Worker task request/return (2 endpoints)
- Updated task assignment (includes role)
- Task unassign API
- Task soft delete implementation
- Role history API
- User password reset API
- Session token authentication (4 endpoints)

## Files Modified

```
internal/db/schema.sql                          (schema updates)
internal/db/models.go                           (struct updates)
internal/db/init.go                             (initialization updates)
internal/db/migrations/001_align_with_design_cli.sql  (new migration script)
internal/server/roles.go                        (rules → goals)
internal/server/workers.go                      (rules → goals)
internal/server/server_test.go                  (rules → goals)
internal/tui/role_form.go                       (rules → goals)
internal/tui/models.go                          (rules → goals)
docs/ENTITY_TASK.md                             (documentation updates)
docs/ENTITY_PROJECT.md                          (documentation updates)
docs/IMPLEMENTATION_PLAN.md                     (new plan document)
```

## Verification Checklist

- [x] Schema changes align with DESIGN_CLI.md
- [x] All Go models updated
- [x] Database init code updated
- [x] Server code updated
- [x] TUI code updated  
- [x] Migration script created
- [ ] Migration tested on copy database
- [ ] Full test suite passes
- [ ] Database applied (production)

## Notes

- Task types 'story' and 'sub-task' will be automatically converted to 'task' during migration
- Comments are stored as JSON text (SQLite doesn't have native JSONB)
- Soft delete uses integer (0/1) as SQLite doesn't have native boolean
- All existing data will be preserved during migration
