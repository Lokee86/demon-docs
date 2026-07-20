package documentpolicy

func BuiltinSchemas() map[string]string {
	return map[string]string{
		"general":  generalSchema,
		"service":  serviceSchema,
		"planning": planningSchema,
		"index":    indexSchema,
	}
}

const generalSchema = `version = 1
name = "general"
description = "Space Rocks-style general documentation"
placeholder = "TODO"
unknown_sections = "manual"
duplicate_sections = "manual"

[document]
title = "{title}"
parent_link = true

[frontmatter]
format = "yaml"

[frontmatter.values]
summary = "TODO"
policy_exempt = false

[[sections]]
id = "purpose"
heading = "Purpose"

[[sections]]
id = "overview"
heading = "Overview"

[[sections]]
id = "related-docs"
heading = "Related docs"
placeholder = "- TODO"

[[sections]]
id = "notes"
heading = "Notes"
`

const serviceSchema = `version = 1
name = "service"
description = "Space Rocks-style service and implementation documentation"
placeholder = "TODO"
unknown_sections = "manual"
duplicate_sections = "manual"

[document]
title = "{title}"
parent_link = true

[frontmatter]
format = "yaml"

[frontmatter.values]
summary = "TODO"
policy_exempt = false

[[sections]]
id = "purpose"
heading = "Purpose"

[[sections]]
id = "overview"
heading = "Overview"

[[sections]]
id = "responsibilities"
heading = "Responsibilities"
placeholder = "- TODO"

[[sections]]
id = "does-not-own"
heading = "Does not own"
placeholder = "- TODO"
aliases = ["Does Not Own", "Does Not Belong"]

[[sections]]
id = "data-ownership"
heading = "Data ownership"

[[sections]]
id = "protocol-api"
heading = "Protocol and API surfaces"

[[sections]]
id = "code-map"
heading = "Code map"

[[sections]]
id = "tests"
heading = "Tests and verification"

[[sections]]
id = "related-docs"
heading = "Related docs"
placeholder = "- TODO"

[[sections]]
id = "notes"
heading = "Notes"
`

const planningSchema = `version = 1
name = "planning"
description = "Space Rocks-style planning documentation"
placeholder = "TODO"
unknown_sections = "manual"
duplicate_sections = "manual"

[document]
title = "{title}"
parent_link = true

[frontmatter]
format = "yaml"

[frontmatter.values]
summary = "TODO"
policy_exempt = false

[[sections]]
id = "purpose"
heading = "Purpose"

[[sections]]
id = "ownership-boundary"
heading = "Ownership Boundary"
aliases = ["Ownership boundary"]

[[sections]]
id = "settled-model"
heading = "Settled Product Model"
aliases = ["Settled product model"]

[[sections]]
id = "system-handoffs"
heading = "System Handoffs"
aliases = ["System handoffs"]

[[sections]]
id = "implementation"
heading = "Implementation Implications"
aliases = ["Implementation implications"]

[[sections]]
id = "open-questions"
heading = "Open Questions"
aliases = ["Open questions"]
optional = true

[[sections]]
id = "related-docs"
heading = "Related Docs"
placeholder = "- TODO"
aliases = ["Related docs"]

[[sections]]
id = "notes"
heading = "Notes"
`

const indexSchema = `version = 1
name = "index"
description = "Space Rocks-style folder index"
placeholder = "TODO"
unknown_sections = "manual"
duplicate_sections = "manual"

[document]
title = "{title}"
parent_link = true

[frontmatter]
format = "yaml"

[frontmatter.values]
summary = "TODO"
policy_exempt = false

[[sections]]
id = "ownership"
heading = "Ownership"

[[sections]]
id = "does-not-belong"
heading = "Does Not Belong"
placeholder = "- TODO"
aliases = ["Does not belong"]

[[sections]]
id = "direct-files"
heading = "Direct Files"
optional = true

[[sections]]
id = "stub-files"
heading = "Stub Files"
optional = true

[[sections]]
id = "direct-folders"
heading = "Direct Folders"
optional = true

[[sections]]
id = "related-docs"
heading = "Related Docs"
placeholder = "- TODO"
aliases = ["Related docs"]

[[sections]]
id = "notes"
heading = "Notes"
`
