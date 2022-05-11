-- name: InsertOrganizationMembership :one
INSERT INTO organization_memberships (
    user_id,
    organization_id
) VALUES (
    pggen.arg('UserID'),
    pggen.arg('OrganizationID')
)
RETURNING *;

-- name: DeleteOrganizationMembership :exec
DELETE
FROM organization_memberships
WHERE
    user_id = pggen.arg('UserID') AND
    organization_id = pggen.arg('OrganizationID')
;
