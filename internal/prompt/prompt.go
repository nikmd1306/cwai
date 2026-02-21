package prompt

import "fmt"

const systemTemplate = `You are a commit message generator. You receive a git diff and output EXACTLY ONE commit message following the Conventional Commits specification.

FORMAT RULES:
1. Format: <type>(<scope>): <description>
   For large changes, add bullet points after an empty line:
   <type>(<scope>): <description>

   - <specific change 1>
   - <specific change 2>

2. Type MUST be one of: feat, fix, refactor, docs, style, test, chore, perf, ci, build
3. Scope is REQUIRED — short, generalizing word covering ALL changes
4. Description: lowercase, imperative mood, no period, max 72 chars first line
5. Bullet points when 3+ distinct changes. No bullet points for trivial changes.
6. Output ONLY the commit message. No explanations, no formatting, no quotes.
7. NEVER use emojis.
8. ALWAYS exactly ONE commit message.
9. Use %s for the commit message.`

const fewShotUserDiff = `diff --git a/internal/auth/jwt.go b/internal/auth/jwt.go
index abc1234..def5678 100644
--- a/internal/auth/jwt.go
+++ b/internal/auth/jwt.go
@@ -1,10 +1,15 @@
-func validateJWT(token string) error {
+func validatePASETO(token string) error {
+    // new validation logic
 }

diff --git a/internal/auth/refresh.go b/internal/auth/refresh.go
new file mode 100644
--- /dev/null
+++ b/internal/auth/refresh.go
@@ -0,0 +1,20 @@
+func rotateRefreshToken(old string) (string, error) {
+    // rotation logic
+}

diff --git a/internal/middleware/auth.go b/internal/middleware/auth.go
index 111aaaa..222bbbb 100644
--- a/internal/middleware/auth.go
+++ b/internal/middleware/auth.go
@@ -5,7 +5,7 @@
-    if err := validateJWT(token); err != nil {
+    if err := validatePASETO(token); err != nil {`

const fewShotAssistantResponse = `refactor(auth): refactored auth system

- replace JWT validation with PASETO tokens
- add refresh token rotation logic
- update middleware to support new token format`

const systemTemplateStructured = `You are a commit message generator. You receive a git diff and produce a structured JSON response that will be assembled into a Conventional Commits message.

COMMIT TYPE DECISION RULES (apply strictly in this order):

RULE 1 - NEW BEHAVIOR: Does the diff add behavior the SYSTEM could not do before?
  -> YES: type = "feat".
  NOTE: Renaming, moving code to new files/modules, or wrapping existing logic in a new type/class does NOT count as new behavior.

RULE 2 - BUG FIX: Does the diff correct wrong/broken behavior?
  -> YES: type = "fix".
  NOTE: If both a bug fix and restructuring are present, prefer "fix".

RULE 3 - RESTRUCTURE: Does the diff reorganize/rename/move existing code without adding new system capabilities?
  -> YES: type = "refactor". New files/types/classes that wrap existing logic count as restructuring.

RULE 4 - Other: docs, test, style, chore, perf, ci, build — use when none of the above apply.

Use %s for the commit message language.
Respond with JSON matching the provided schema. Fill boolean guard fields BEFORE selecting the type.`

const structuredFewShotUserDiff1 = `diff --git a/src/handlers/user.py b/src/handlers/user.py
--- a/src/handlers/user.py
+++ b/src/handlers/user.py
@@ -10,15 +10,5 @@
-def create_user(data):
-    if not data.get("email"):
-        raise ValueError("email required")
-    if len(data.get("password", "")) < 8:
-        raise ValueError("password too short")
-    if not re.match(r"[^@]+@[^@]+\.[^@]+", data["email"]):
-        raise ValueError("invalid email format")
-    return save_user(data)
+def create_user(data):
+    validate_user_input(data)
+    return save_user(data)

diff --git a/src/validators/user_validator.py b/src/validators/user_validator.py
new file mode 100644
--- /dev/null
+++ b/src/validators/user_validator.py
@@ -0,0 +1,10 @@
+def validate_user_input(data):
+    if not data.get("email"):
+        raise ValueError("email required")
+    if len(data.get("password", "")) < 8:
+        raise ValueError("password too short")
+    if not re.match(r"[^@]+@[^@]+\.[^@]+", data["email"]):
+        raise ValueError("invalid email format")`

const structuredFewShotResponse1 = `{"changes_summary":"Validation logic extracted from user handler into a dedicated validator module.","introduces_new_behavior":false,"fixes_broken_behavior":false,"restructures_only":true,"type_reasoning":"Code moved from handler to new file. No new system capability added. restructures_only=true -> refactor.","type":"refactor","scope":"validation","description":"extract user validation into dedicated module","bullet_points":null}`

const structuredFewShotUserDiff2 = `diff --git a/src/middleware/rateLimiter.js b/src/middleware/rateLimiter.js
new file mode 100644
--- /dev/null
+++ b/src/middleware/rateLimiter.js
@@ -0,0 +1,8 @@
+const rateLimit = require("express-rate-limit");
+
+module.exports = rateLimit({
+  windowMs: 15 * 60 * 1000,
+  max: 100,
+  message: { error: "Too many requests" }
+});

diff --git a/src/routes/api.js b/src/routes/api.js
--- a/src/routes/api.js
+++ b/src/routes/api.js
@@ -1,6 +1,8 @@
 const express = require("express");
+const rateLimiter = require("../middleware/rateLimiter");
 const router = express.Router();
+router.use(rateLimiter);

diff --git a/config/default.json b/config/default.json
--- a/config/default.json
+++ b/config/default.json
@@ -3,5 +3,9 @@
   "port": 3000,
+  "rateLimit": {
+    "windowMs": 900000,
+    "max": 100
+  }
 }`

const structuredFewShotResponse2 = `{"changes_summary":"Added rate limiting middleware to API routes with configurable window and max requests.","introduces_new_behavior":true,"fixes_broken_behavior":false,"restructures_only":false,"type_reasoning":"Rate limiting is a new capability the system did not have before. introduces_new_behavior=true -> feat.","type":"feat","scope":"api","description":"add rate limiting to api endpoints","bullet_points":["add rate limiter middleware with 100 req/15min window","apply rate limiting to all api routes","add rate limit settings to default config"]}`

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func BuildMessages(language, diff string, structuredOutput bool) []Message {
	if structuredOutput {
		system := fmt.Sprintf(systemTemplateStructured, language)
		return []Message{
			{Role: "system", Content: system},
			{Role: "user", Content: structuredFewShotUserDiff1},
			{Role: "assistant", Content: structuredFewShotResponse1},
			{Role: "user", Content: structuredFewShotUserDiff2},
			{Role: "assistant", Content: structuredFewShotResponse2},
			{Role: "user", Content: diff},
		}
	}

	system := fmt.Sprintf(systemTemplate, language)
	return []Message{
		{Role: "system", Content: system},
		{Role: "user", Content: fewShotUserDiff},
		{Role: "assistant", Content: fewShotAssistantResponse},
		{Role: "user", Content: diff},
	}
}
