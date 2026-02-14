# Documentation Review - Contradictions & Recommendations

**Date:** 2026-02-09
**Reviewer:** Automated documentation analysis
**Scope:** All markdown documentation files

---

## Critical Issues

### 1. Binary Name Inconsistency

**Issue:** Documentation references multiple binary names inconsistently.

**Current State:**
- Codebase uses `sf` as the binary name
- README.md line 107: References `cd task`
- CONTRIBUTING.md line 18: References repository clone as `cd task`
- Multiple references to "task" CLI throughout older docs

**Recommendation:**
- ✅ **Already Fixed:** Binary is correctly named `sf` in code
- ❌ **Needs Fix:** Update README.md line 107 to clarify: `cd tsf` (repository name)
- ❌ **Needs Fix:** Update CONTRIBUTING.md line 18 to use correct repository name
- ❌ **Needs Fix:** Global find/replace `./task` → `./sf` in documentation (already done in code)

---

### 2. Go Version Requirement Mismatch

**Issue:** Two different Go versions are specified.

**Contradictions:**
- README.md line 61: "**Go 1.24+**"
- CONTRIBUTING.md line 9: "Go 1.23 or later"

**Recommendation:**
- Standardize on **Go 1.24+** (matches README, more conservative)
- Update CONTRIBUTING.md line 9 to match README.md
- Or verify minimum version and document the tested/supported version explicitly

---

### 3. Environment Variable Naming Inconsistency

**Issue:** Multiple names for the same server URL configuration.

**Contradictions:**
- README.md line 158: `SF_SERVER_URL=http://localhost:8080`
- USER_GUIDE.md lines 329, 804: Uses both `SF_SERVER_URL` AND `SF_URL`
- DESIGN_CLI.md line 38: References `$SF_URL`
- Actual code may use different names

**Recommendation:**
- **Choose ONE:** Either `SF_URL` or `SF_SERVER_URL`
- Update all documentation to use the chosen variable consistently
- Add backward compatibility note if changing
- Suggested: Use `SF_URL` (shorter, matches other vars like `SF_USERNAME`)

---

### 4. Config Command Flag Inconsistency

**Issue:** Different flag names for the same option.

**Contradictions:**
- README.md line 261: `sf config set -key KEY -val VALUE`
- DESIGN_CLI.md line 118: `sf config set -key KEY -value VALUE`

**Recommendation:**
- Verify actual implementation in code
- Standardize on `-value` (more explicit and clear)
- Update all documentation examples
- OR support both `-val` and `-value` as aliases

---

### 5. Flag Convention Inconsistency

**Issue:** Mixed use of single-dash and double-dash for long options.

**Observations:**
- DESIGN_CLI.md line 32: "Use single-hyphen for options, never double (except for --force)"
- README.md line 122: Shows `--force` and `--password`
- Most examples: `-username`, `-password` (single dash)

**Recommendation:**
- Current approach is reasonable: single-dash for all flags
- Exception: `--force` and `--populate` as special flags
- Document the convention clearly:
  - Single dash for all named parameters: `-key`, `-value`, `-username`
  - Double dash for boolean flags: `--force`, `--populate`, `--json`
- Update DESIGN_CLI.md to clarify this convention

---

## Documentation Quality Issues

### 6. Default Password Documentation Ambiguity

**Issue:** Conflicting information about default passwords.

**Contradictions:**
- README.md lines 131-133: Says passwords are random OR specified via `--password`
- USER_GUIDE.md line 47: States default admin password is `admin123`

**Recommendation:**
- Clarify that `admin123` is ONLY the default when using `--password admin123`
- Document clearly:
  - **Without --password:** Random passwords printed to stdout
  - **With --password:** All users get the specified password
  - **Security:** Change defaults immediately in production
- Add warning: "Never use default passwords in production"

---

### 7. Database Path Inconsistency

**Issue:** Old references to `task.db` still exist.

**Observations:**
- Most docs correctly use `~/.config/sf/sf.db`
- Old references may exist to `~/.config/task/task.db`

**Recommendation:**
- Global search for `task.db` in documentation
- Replace all with `sf.db`
- Ensure quickstart.sh uses correct path (already fixed in code)

---

### 8. Project Structure Documentation Outdated

**Issue:** README.md shows outdated directory structure.

**Current Doc (README.md line 306-326):**
```
├── cmd/
│   └── task-unified/      # Single unified binary (task)
```

**Recommendation:**
- Verify actual project structure
- Update to reflect current layout (main.go in root?)
- Remove references to `cmd/task-unified` if no longer accurate

---

### 9. Missing Help Command Documentation

**Issue:** Help command usage not documented in user-facing docs.

**Observations:**
- DESIGN_CLI.md documents: `sf help <command>` and `sf <command> -h`
- USER_GUIDE.md does not mention help command usage
- Help functions exist and work correctly (recently tested)

**Recommendation:**
- Add section to USER_GUIDE.md:
  ```markdown
  ### Getting Help

  View help for any command:
  ```bash
  # Show general help
  sf help

  # Show command-specific help
  sf help project
  sf help task
  sf help user
  sf help role
  sf help config

  # Alternative syntax
  sf project -h
  sf task -h
  ```
  ```

---

### 10. Test Coverage Documentation

**Issue:** Testing documentation incomplete.

**Observations:**
- README.md mentions test commands
- No mention of the comprehensive navigation tests recently added
- No documentation of test coverage requirements

**Recommendation:**
- Document test requirements:
  - All navigation links must have tests
  - No JavaScript errors in any screen
  - Minimum coverage threshold (if established)
- Add section about E2E test philosophy:
  - Tests click every navigation link
  - Tests verify no console errors
  - Tests verify graceful degradation

---

## Recommendations for Improvement

### 11. Add Troubleshooting Section to README

**Current:** Troubleshooting only in USER_GUIDE.md

**Recommendation:**
- Add quick troubleshooting section to README.md:
  - Port already in use
  - Database not initialized
  - Authentication failures
  - Link to full USER_GUIDE.md for details

---

### 12. Add Architecture Diagram Update

**Issue:** ASCII architecture diagram may not reflect current state.

**Recommendation:**
- Verify architecture diagram (README.md lines 31-55) is accurate
- Add note about single binary containing all components
- Clarify that server serves both API and web UI

---

### 13. Standardize Example Formatting

**Issue:** Mixed formatting styles for code examples.

**Recommendation:**
- Use consistent formatting:
  - All bash commands in ```bash blocks
  - All outputs in ```text or ```json blocks
  - Consistent use of `$` prompt or no prompt
- Pick one style and apply everywhere

---

### 14. Add Quick Reference Card

**Recommendation:**
- Create `QUICKREF.md` with:
  - Most common commands
  - Environment variables
  - Default ports
  - File locations
  - Emergency procedures

---

### 15. Document Recent Changes

**Changes Not Yet Documented:**
- Users screen added (recently implemented)
- Comprehensive navigation tests added
- JavaScript error handling improvements
- Robust fallback behavior for missing screens

**Recommendation:**
- Add CHANGELOG.md documenting:
  - Recent bug fixes
  - New features
  - Breaking changes
  - Migration notes

---

## Priority Action Items

### High Priority (Do Immediately)

1. ✅ **Binary name consistency** - Update all `task` references to `sf`
2. ✅ **Go version standardization** - Pick 1.23 or 1.24 and document
3. ✅ **Environment variable naming** - Standardize on `SF_URL` or `SF_SERVER_URL`
4. ✅ **Default password clarification** - Document random vs fixed passwords

### Medium Priority (Do Soon)

5. ⚠️ **Config command flags** - Standardize `-val` vs `-value`
6. ⚠️ **Help command docs** - Add to USER_GUIDE.md
7. ⚠️ **Project structure** - Update README.md structure diagram
8. ⚠️ **Database paths** - Remove all `task.db` references

### Low Priority (Nice to Have)

9. 📝 **Quick reference** - Create QUICKREF.md
10. 📝 **Changelog** - Create CHANGELOG.md
11. 📝 **Troubleshooting** - Add to README.md
12. 📝 **Testing docs** - Document test philosophy and requirements

---

## Verification Checklist

After addressing these recommendations:

- [ ] All documentation uses `sf` (not `task`) for binary name
- [ ] Go version is consistent (1.24+)
- [ ] Environment variables are standardized
- [ ] Config command flags are consistent
- [ ] Default passwords are clearly documented
- [ ] Help command usage is documented
- [ ] Project structure diagram matches reality
- [ ] All examples use correct paths and commands
- [ ] Test coverage and philosophy documented
- [ ] CHANGELOG exists for tracking changes

---

## Notes

This review was conducted after:
- Renaming binary from `task` to `sf`
- Updating Playwright tests and configurations
- Adding comprehensive navigation tests
- Implementing robust JavaScript error handling
- Adding Users screen to web UI

Some documentation may lag behind code changes, which is normal but should be addressed to prevent confusion.

---

**End of Report**
