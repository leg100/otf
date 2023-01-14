-- name: InsertOrganizationMembership :exec
INSERT INTO organization_memberships (
    user_id,
    organization_id
) VALUES (
    pggen.arg('user_id'),
    pggen.arg('organization_id')
)
;

-- name: DeleteOrganizationMembership :one
DELETE
FROM organization_memberships
WHERE user_id         = pggen.arg('user_id')
AND   organization_id = pggen.arg('organization_id')
RETURNING user_id
;
