-- name: InsertOrganizationMembership :exec
INSERT INTO organization_memberships (
    user_id,
    organization_name
) VALUES (
    pggen.arg('user_id'),
    pggen.arg('organization_name')
)
;

-- name: DeleteOrganizationMembership :one
DELETE
FROM organization_memberships
WHERE user_id           = pggen.arg('user_id')
AND   organization_name = pggen.arg('organization_name')
RETURNING user_id
;
