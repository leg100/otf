-- name: InsertOrganizationMembership :exec
INSERT INTO organization_memberships (
    username,
    organization_name
) VALUES (
    pggen.arg('username'),
    pggen.arg('organization_name')
)
;

-- name: DeleteOrganizationMembership :one
DELETE
FROM organization_memberships
WHERE username          = pggen.arg('username')
AND   organization_name = pggen.arg('organization_name')
RETURNING username
;
