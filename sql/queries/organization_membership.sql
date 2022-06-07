-- name: InsertOrganizationMembership :exec
INSERT INTO organization_memberships (
    user_id,
    organization_id
) VALUES (
    pggen.arg('UserID'),
    pggen.arg('OrganizationID')
)
;

-- name: DeleteOrganizationMembership :one
DELETE
FROM organization_memberships
WHERE
    user_id = pggen.arg('UserID') AND
    organization_id = pggen.arg('OrganizationID')
RETURNING user_id
;
