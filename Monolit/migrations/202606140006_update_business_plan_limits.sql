-- +goose Up
UPDATE plans
SET company_limit = v.company_limit,
    departments_per_company_limit = v.departments_per_company_limit,
    instructions_per_department_limit = v.instructions_per_department_limit,
    updated_at = now()
FROM (
    VALUES
        ('business_start', 1, 5, 5),
        ('business_plus', 1, 10, 7),
        ('business_pro', 3, 15, 10)
) AS v(code, company_limit, departments_per_company_limit, instructions_per_department_limit)
WHERE plans.code = v.code;

-- +goose Down
SELECT 1;
