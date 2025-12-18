-- name: CreateParticipant :one
INSERT INTO participants (
    name, email, phone_number, problem_statement
) VALUES (
    ?, ?, ?, ?
)
RETURNING *;
