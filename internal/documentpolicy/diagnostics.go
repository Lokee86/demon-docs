package documentpolicy

const (
	diagnosticSchemaLoadError          = "format.schema_load_error"
	diagnosticFrontmatterParseError    = "format.frontmatter_parse_error"
	diagnosticSchemaSelectionError     = "format.schema_selection_error"
	diagnosticDocumentSchemaLoadError  = "format.document_schema_load_error"
	diagnosticDocumentIdentityMismatch = "format.document_schema_identity_mismatch"
	diagnosticDocumentSharedMismatch   = "format.document_schema_shared_mismatch"
	diagnosticSchemaSnapshotLoadError  = "format.schema_snapshot_load_error"
	diagnosticSchemaSnapshotMissing    = "format.schema_snapshot_missing"
	diagnosticDocumentSchemaInvalid    = "format.document_schema_invalidated"
	diagnosticEffectiveSchemaInvalid   = "format.effective_schema_invalid"
	diagnosticAmbiguousSection         = "format.ambiguous_section"
	diagnosticUnknownSection           = "format.unknown_section"
	diagnosticDuplicateSection         = "format.duplicate_section"
	diagnosticSectionParent            = "format.section_parent"
	diagnosticRequiredSectionMissing   = "format.required_section_missing"
	diagnosticSectionRenamed           = "format.section_renamed"
	diagnosticAliasCanonicalization    = "format.alias_canonicalization"
	diagnosticHeadingLevel             = "format.heading_level"
	diagnosticSectionOrder             = "format.section_order"
)
