package repository

const queryScreenByKey = `
	SELECT
		si.id, si.screen_key, si.template_id, si.name, si.description,
		si.slot_data,
		si.scope, si.required_permission, si.handler_key, si.is_active,
		si.created_at, si.updated_at,
		st.id, st.pattern, st.name, st.description, st.version,
		st.definition, st.is_active, st.created_at, st.updated_at
	FROM ui_config.screen_instances si
	JOIN ui_config.screen_templates st ON st.id = si.template_id
	WHERE si.screen_key = ?
	  AND si.is_active = true`

const queryScreensByResourceKey = `
	SELECT
		si.id, si.screen_key, si.template_id, si.name, si.description,
		si.slot_data,
		si.scope, si.required_permission, si.handler_key, si.is_active,
		si.created_at, si.updated_at,
		st.id, st.pattern, st.name, st.description, st.version,
		st.definition, st.is_active, st.created_at, st.updated_at
	FROM ui_config.resource_screens rs
	JOIN ui_config.screen_instances si ON si.screen_key = rs.screen_key
	JOIN ui_config.screen_templates st ON st.id = si.template_id
	WHERE rs.resource_key = ?
	  AND rs.is_active = true
	  AND si.is_active = true
	ORDER BY rs.sort_order`

const queryUpsertPreferences = `
	INSERT INTO ui_config.screen_instances (screen_key, slot_data, updated_at)
	VALUES (?, ?, NOW())
	ON CONFLICT (screen_key)
	DO UPDATE SET slot_data = EXCLUDED.slot_data, updated_at = NOW()`
