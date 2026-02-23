package repository

const queryAvgAttemptScoreBySchool = `
	SELECT AVG(aa.percentage)
	FROM assessment.assessment_attempt aa
	JOIN assessment.assessment a ON a.id = aa.assessment_id
	JOIN content.materials m ON m.id = a.material_id
	WHERE aa.status = 'completed'
	  AND m.school_id = ?
	  AND m.deleted_at IS NULL`

const queryMaterialAttemptStats = `
	SELECT
		COUNT(*) AS total_attempts,
		COALESCE(AVG(aa.percentage), 0) AS average_score,
		COUNT(DISTINCT aa.student_id) AS unique_students
	FROM assessment.assessment_attempt aa
	JOIN assessment.assessment a ON a.id = aa.assessment_id
	WHERE a.material_id = ?
	  AND aa.status = 'completed'`

const queryMaterialCompletionRate = `
	SELECT
		CASE WHEN COUNT(*) > 0
			THEN CAST(SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) AS FLOAT) / COUNT(*)
			ELSE 0
		END
	FROM content.progress
	WHERE material_id = ?`
