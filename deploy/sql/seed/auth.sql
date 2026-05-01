BEGIN;

-- Permission data is projected from admin into auth.casbin_rules.
-- Keep this file as a no-op so the seed flow can retain the auth step
-- without reintroducing stale permission master tables in the auth DB.

COMMIT;
